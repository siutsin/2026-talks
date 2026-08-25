package main

import (
	"math/rand"
)

// Strategy defines the parameters for the worker pool and its evolutionary search.
type Strategy struct {
	MaxWorkers        int32   // Maximum concurrent fetchers allowed
	BatchSize         int32   // How many URLs a worker grabs at once
	MaxRetries        int32   // How many times to retry a failed URL
	BackoffMultiplier float64 // Multiplier for delay between retries
	Kp                float64 // Proportional gain for the PID backpressure reflex (error rate)
}

// Clone returns a deep copy of the Strategy.
func (s Strategy) Clone() Strategy {
	return Strategy{
		MaxWorkers:        s.MaxWorkers,
		BatchSize:         s.BatchSize,
		MaxRetries:        s.MaxRetries,
		BackoffMultiplier: s.BackoffMultiplier,
		Kp:                s.Kp,
	}
}

// Mutate returns a mutated version of the Strategy for the GA evolution.
func Mutate(s Strategy) Strategy {
	m := s.Clone()
	m.MaxWorkers += int32(rand.Intn(7) - 3)
	if m.MaxWorkers < 1 {
		m.MaxWorkers = 1
	}
	if m.MaxWorkers > 30 {
		m.MaxWorkers = 30
	} // Safety pool limit cap
	m.BatchSize += int32(rand.Intn(3) - 1)
	if m.BatchSize < 1 {
		m.BatchSize = 1
	}
	m.MaxRetries += int32(rand.Intn(3) - 1)
	if m.MaxRetries < 0 {
		m.MaxRetries = 0
	}
	if m.MaxRetries > 5 {
		m.MaxRetries = 5
	}
	m.BackoffMultiplier += (rand.Float64() * 0.5) - 0.25
	if m.BackoffMultiplier < 1.0 {
		m.BackoffMultiplier = 1.0
	}
	if m.BackoffMultiplier > 3.0 {
		m.BackoffMultiplier = 3.0
	}
	return m
}
