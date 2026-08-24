package main

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestAtomicPIDCompute(t *testing.T) {
	pid := &AtomicPID{Kp: math.Float64bits(10.0)}

	// Test 0 error rate
	adjust := pid.Compute(0.0)
	if adjust != 0.0 {
		t.Errorf("Expected 0.0 adjustment for 0.0 error rate, got %f", adjust)
	}

	// Test positive error rate (e.g., 25% error)
	adjust = pid.Compute(0.25)
	if adjust != 2.5 { // 10.0 * 0.25
		t.Errorf("Expected 2.5 adjustment for 0.25 error rate, got %f", adjust)
	}
}

func TestEvaluateFitness(t *testing.T) {
	telemetry := &Telemetry{}
	strategy := Strategy{}
	window := 5 * time.Second

	// Case 1: 0 total requests
	score := Evaluate(telemetry, window, strategy)
	if score != 0 {
		t.Errorf("Expected 0 score for 0 requests, got %f", score)
	}

	// Case 2: 100% success, 1MB read over 5s
	// Mbps = (1024 * 1024 / (1024*1024)) / 5s = 0.2 Mbps
	// Reliability = 1.0^2 = 1.0
	// Expected score = 0.2
	telemetry.Record(100*time.Millisecond, 1024*1024, nil)
	score = Evaluate(telemetry, window, strategy)
	
	expectedMbps := 0.2
	if math.Abs(score-expectedMbps) > 0.001 {
		t.Errorf("Expected score around %.3f, got %f", expectedMbps, score)
	}

	// Case 3: 50% success (1 success, 1 failure)
	// Reliability = 0.5^2 = 0.25
	// Total bytes is still 1MB, so mbps = 0.2
	// Expected score = 0.2 * 0.25 = 0.05
	telemetry.Record(100*time.Millisecond, 0, errors.New("fake error"))
	score = Evaluate(telemetry, window, strategy)
	
	expectedScore := 0.05
	if math.Abs(score-expectedScore) > 0.001 {
		t.Errorf("Expected score around %.3f, got %f", expectedScore, score)
	}
}

func TestTelemetryReset(t *testing.T) {
	telemetry := &Telemetry{}
	telemetry.Record(100*time.Millisecond, 1024, nil)
	telemetry.Record(100*time.Millisecond, 0, errors.New("fake error"))

	telemetry.Reset()

	if telemetry.Successes != 0 || telemetry.Failures != 0 || telemetry.BytesRead != 0 || telemetry.TotalLat != 0 {
		t.Errorf("Telemetry was not fully reset: %+v", telemetry)
	}
	
	if telemetry.CumulativeBytes != 1024 {
		t.Errorf("CumulativeBytes should not be reset, got %d", telemetry.CumulativeBytes)
	}
}