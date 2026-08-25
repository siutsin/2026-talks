package main

import (
	"testing"
)

func TestStrategyClone(t *testing.T) {
	s1 := Strategy{
		MaxWorkers:        10,
		BatchSize:         5,
		MaxRetries:        3,
		BackoffMultiplier: 2.0,
		Kp:                1.5,
	}

	s2 := s1.Clone()
	
	if s2.MaxWorkers != s1.MaxWorkers || 
	   s2.BatchSize != s1.BatchSize || 
	   s2.MaxRetries != s1.MaxRetries ||
	   s2.BackoffMultiplier != s1.BackoffMultiplier ||
	   s2.Kp != s1.Kp {
		t.Errorf("Clone did not match original. Original: %+v, Clone: %+v", s1, s2)
	}

	// Ensure modifying the clone doesn't affect the original
	s2.MaxWorkers = 99
	if s1.MaxWorkers == 99 {
		t.Errorf("Clone is not a deep copy. Modifying clone modified original.")
	}
}

func TestStrategyMutateBounds(t *testing.T) {
	s := Strategy{
		MaxWorkers:        1,
		BatchSize:         1,
		MaxRetries:        0,
		BackoffMultiplier: 1.0,
		Kp:                1.5,
	}

	// Mutate enough times to ensure it stays within safety bounds
	for i := 0; i < 100; i++ {
		m := Mutate(s)
		
		if m.MaxWorkers < 1 || m.MaxWorkers > 30 {
			t.Errorf("MaxWorkers bound exceeded: %d", m.MaxWorkers)
		}
		if m.BatchSize < 1 {
			t.Errorf("BatchSize bound exceeded: %d", m.BatchSize)
		}
		if m.MaxRetries < 0 || m.MaxRetries > 5 {
			t.Errorf("MaxRetries bound exceeded: %d", m.MaxRetries)
		}
		if m.BackoffMultiplier < 1.0 || m.BackoffMultiplier > 3.0 {
			t.Errorf("BackoffMultiplier bound exceeded: %f", m.BackoffMultiplier)
		}
		
		s = m // Evolve forward
	}
}