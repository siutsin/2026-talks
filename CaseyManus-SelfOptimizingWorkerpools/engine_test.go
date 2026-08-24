package main

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// MockFetcher simulates network fetching for tests.
type MockFetcher struct {
	ShouldFail bool
	Calls      int
}

func (m *MockFetcher) Fetch(url string) ([]string, int, error) {
	m.Calls++
	if m.ShouldFail {
		return nil, 0, errors.New("mock network error")
	}
	return []string{"http://mock.com/link1"}, 100, nil
}

func TestEngineAdjustWorkers(t *testing.T) {
	e := NewEngine(&MockFetcher{})
	
	// Initial state is 0 workers since Start hasn't been called
	if atomic.LoadInt32(&e.ActiveWorkers) != 0 {
		t.Errorf("Expected 0 workers, got %d", e.ActiveWorkers)
	}

	// Scale up
	e.AdjustWorkers(5)
	if atomic.LoadInt32(&e.ActiveWorkers) != 5 {
		t.Errorf("Expected 5 workers, got %d", e.ActiveWorkers)
	}

	// Scale down
	e.AdjustWorkers(2)
	if atomic.LoadInt32(&e.ActiveWorkers) != 2 {
		t.Errorf("Expected 2 workers, got %d", e.ActiveWorkers)
	}
}

func TestEngineProcessURLWithRetries(t *testing.T) {
	mock := &MockFetcher{ShouldFail: true}
	e := NewEngine(mock)
	s := e.GetStrategy()
	
	// Set retries to 2 (total 3 attempts)
	s.MaxRetries = 2
	
	// Fast backoff for tests
	s.BackoffMultiplier = 1.0
	
	success := e.processURLWithRetries("http://test.com", s)
	
	if success {
		t.Errorf("Expected processURLWithRetries to fail")
	}
	
	if mock.Calls != 3 {
		t.Errorf("Expected 3 fetch attempts (1 initial + 2 retries), got %d", mock.Calls)
	}
	
	if atomic.LoadUint64(&e.Telemetry.Failures) != 3 {
		t.Errorf("Expected 3 failures recorded in telemetry, got %d", e.Telemetry.Failures)
	}

	// Test Success Path
	mock.ShouldFail = false
	mock.Calls = 0
	
	success = e.processURLWithRetries("http://test.com", s)
	if !success {
		t.Errorf("Expected processURLWithRetries to succeed")
	}
	if mock.Calls != 1 {
		t.Errorf("Expected exactly 1 fetch attempt, got %d", mock.Calls)
	}
	if atomic.LoadUint64(&e.Telemetry.Successes) != 1 {
		t.Errorf("Expected 1 success recorded in telemetry, got %d", e.Telemetry.Successes)
	}
}

func TestEngineComputePIDThrottle(t *testing.T) {
	e := NewEngine(&MockFetcher{})
	
	// Simulate 0% error rate
	throttle := e.computePIDThrottle(0.0)
	if throttle != 0 {
		t.Errorf("Expected 0 throttle for 0 error rate, got %v", throttle)
	}

	// Simulate 50% error rate (EMA = 0.5)
	// Kp is 20.0 (from NewEngine default)
	// Adjust = 20.0 * 0.5 = 10.0
	// Throttle should be 10 seconds
	throttle = e.computePIDThrottle(0.5)
	if throttle != 10 * time.Second {
		t.Errorf("Expected 10s throttle for 0.5 error rate, got %v", throttle)
	}

	// Simulate >100% massive spike, verify clamping
	throttle = e.computePIDThrottle(2.0)
	if throttle != 30 * time.Second { // Clamped at 30s
		t.Errorf("Expected max throttle of 30s, got %v", throttle)
	}
}

func TestEngineUIModes(t *testing.T) {
	e := NewEngine(&MockFetcher{})
	
	if e.GetUIMode() != "INITIALIZING" {
		t.Errorf("Expected initial UI mode 'INITIALIZING', got %s", e.GetUIMode())
	}

	e.SetUIMode("TEST_MODE")
	if e.GetUIMode() != "TEST_MODE" {
		t.Errorf("Expected UI mode 'TEST_MODE', got %s", e.GetUIMode())
	}
}

func TestEngineStrategyUpdate(t *testing.T) {
	e := NewEngine(&MockFetcher{})
	
	newStrategy := Strategy{
		MaxWorkers:        10,
		BatchSize:         5,
		MaxRetries:        2,
		BackoffMultiplier: 2.0,
		Kp:                5.0,
	}

	e.updateStrategy(newStrategy)

	s := e.GetStrategy()
	if s.MaxWorkers != 10 || s.BatchSize != 5 {
		t.Errorf("Strategy was not updated correctly: %+v", s)
	}

	if atomic.LoadInt32(&e.ActiveWorkers) != 10 {
		t.Errorf("Active workers were not adjusted during strategy update. Expected 10, got %d", e.ActiveWorkers)
	}
}