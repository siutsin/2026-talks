package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// Engine manages the worker pool, telemetry, and evolutionary search.
type Engine struct {
	Queue           *URLQueue
	Telemetry       *Telemetry
	Fetcher         Fetcher
	PID             *AtomicPID
	mu              sync.RWMutex
	CurrentStrategy Strategy

	ActiveWorkers    int32
	WorkerIDSequence int32
	CurrentBatch     int32
	UIMode           string

	GenWindow time.Duration
}

// NewEngine creates a new Engine with default values.
func NewEngine(f Fetcher) *Engine {
	strategy := Strategy{
		MaxWorkers:        4,
		BatchSize:         1,
		MaxRetries:        1,
		BackoffMultiplier: 1.5,
		Kp:                20.0, // Higher gain for error rate reaction
	}
	return &Engine{
		Queue:           NewURLQueue(),
		Telemetry:       &Telemetry{},
		Fetcher:         f,
		PID:             &AtomicPID{Kp: math.Float64bits(strategy.Kp)},
		CurrentStrategy: strategy,
		CurrentBatch:    strategy.BatchSize,
		UIMode:          "INITIALIZING",
		GenWindow:       5 * time.Second,
	}
}

// GetStrategy returns a copy of the current strategy.
func (e *Engine) GetStrategy() Strategy {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.CurrentStrategy
}

// GetUIMode returns the current UI mode string.
func (e *Engine) GetUIMode() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.UIMode
}

// SetUIMode updates the UI mode string.
func (e *Engine) SetUIMode(mode string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.UIMode = mode
}

// Start begins the background processes for the engine.
func (e *Engine) Start() {
	e.Queue.Push([]string{
		"https://en.wikipedia.org/wiki/Artificial_intelligence",
		"https://en.wikipedia.org/wiki/Go_(programming_language)",
		"https://en.wikipedia.org/wiki/Control_theory",
	})

	go e.seederLoop()
	e.AdjustWorkers(e.GetStrategy().MaxWorkers)
	go e.evolutionLoop()
}

func (e *Engine) seederLoop() {
	for {
		if e.Queue.Len() < 5 {
			e.Queue.Push([]string{fmt.Sprintf("https://en.wikipedia.org/wiki/Special:Random?r=%d", rand.Intn(100000))})
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// AdjustWorkers changes the number of active workers to match the target.
func (e *Engine) AdjustWorkers(target int32) {
	for {
		current := atomic.LoadInt32(&e.ActiveWorkers)
		if current == target {
			break
		}
		if current < target {
			if atomic.CompareAndSwapInt32(&e.ActiveWorkers, current, current+1) {
				id := atomic.AddInt32(&e.WorkerIDSequence, 1) - 1
				go e.worker(id)
			}
		} else {
			_ = atomic.CompareAndSwapInt32(&e.ActiveWorkers, current, current-1)
		}
	}
}

func (e *Engine) worker(myID int32) {
	var currentEMA float64
	alpha := 0.2 // Smoothing factor for EMA (5-second roughly depends on loop rate)
	
	for {
		if myID >= atomic.LoadInt32(&e.ActiveWorkers) {
			return
		}
		
		s := e.GetStrategy()
		batchSize := int(atomic.LoadInt32(&e.CurrentBatch))
		urls := e.Queue.PopBatch(batchSize)
		
		if len(urls) == 0 {
			time.Sleep(250 * time.Millisecond)
			continue
		}
		
		for _, url := range urls {
			if myID >= atomic.LoadInt32(&e.ActiveWorkers) {
				return
			}

			success := e.processURLWithRetries(url, s)
			
			// Update Error Rate EMA (0.0 for success, 1.0 for error)
			errorVal := 0.0
			if !success {
				errorVal = 1.0
			}
			currentEMA = (alpha * errorVal) + ((1.0 - alpha) * currentEMA)
			
			pidThrottle := e.computePIDThrottle(currentEMA)
			if pidThrottle > 0 {
				time.Sleep(pidThrottle)
			}
		}
	}
}

func (e *Engine) processURLWithRetries(url string, s Strategy) bool {
	retries := 0
	baseDelay := 100 * time.Millisecond
	for {
		start := time.Now()
		links, bytes, err := e.Fetcher.Fetch(url)
		duration := time.Since(start)
		e.Telemetry.Record(duration, bytes, err)

		if err == nil {
			if len(links) > 0 {
				e.Queue.Push(links)
			}
			return true // Success
		}
		
		if int32(retries) >= s.MaxRetries {
			return false // Final failure
		}
		
		retries++
		time.Sleep(baseDelay)
		baseDelay = time.Duration(float64(baseDelay) * s.BackoffMultiplier)
	}
}

func (e *Engine) computePIDThrottle(errorRateEMA float64) time.Duration {
	// PID Output is based on error rate. Target is 0.0 (0% error).
	adjust := e.PID.Compute(errorRateEMA)
	if adjust > 0 {
		pidThrottle := time.Duration(adjust * float64(time.Second))
		if pidThrottle > 30*time.Second {
			pidThrottle = 30 * time.Second
		}
		return pidThrottle
	}
	return 0
}

func (e *Engine) evolutionLoop() {
	bestScore := 0.0
	bestStrategy := e.GetStrategy().Clone()
	e.SetUIMode("DARWINIAN SEARCH")

	for {
		time.Sleep(e.GenWindow)
		strategy := e.GetStrategy()
		score := Evaluate(e.Telemetry, e.GenWindow, strategy)
		succ := atomic.LoadUint64(&e.Telemetry.Successes)
		fail := atomic.LoadUint64(&e.Telemetry.Failures)
		total := succ + fail

		// Elitism check
		if score > bestScore && fail == 0 && total > 0 {
			bestScore = score
			bestStrategy = strategy.Clone()
		}

		nextStrategy := e.selectNextStrategy(bestStrategy)

		e.updateStrategy(nextStrategy)
		e.Telemetry.Reset()
	}
}

func (e *Engine) selectNextStrategy(bestStrategy Strategy) Strategy {
	e.SetUIMode("DARWINIAN SEARCH")
	if rand.Float64() < 0.75 {
		return Mutate(bestStrategy)
	}
	return bestStrategy.Clone()
}

func (e *Engine) updateStrategy(s Strategy) {
	e.mu.Lock()
	e.CurrentStrategy = s
	e.mu.Unlock()
	atomic.StoreInt32(&e.CurrentBatch, s.BatchSize)
	atomic.StoreUint64(&e.PID.Kp, math.Float64bits(s.Kp))
	e.AdjustWorkers(s.MaxWorkers)
}

