package main

import (
	"sync/atomic"
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestUIRefresh(t *testing.T) {
	// 1. Setup Mock Engine
	engine := NewEngine(&MockFetcher{})
	
	// Inject some fake telemetry and state
	atomic.StoreInt32(&engine.ActiveWorkers, 7)
	atomic.StoreInt32(&engine.CurrentBatch, 3)
	engine.SetUIMode("TESTING_MODE")
	
	// Fake some bytes and successes
	// 2.5 MB = 2.5 * 1024 * 1024 = 2621440 bytes
	atomic.StoreUint64(&engine.Telemetry.CumulativeBytes, 2621440)
	atomic.StoreUint64(&engine.Telemetry.CumulativeSuccesses, 95)
	atomic.StoreUint64(&engine.Telemetry.CumulativeFailures, 5)
	
	// Add 3 fake items to the queue
	engine.Queue.Push([]string{"u1", "u2", "u3"})

	// 2. Initialize UI using Fyne's headless test app
	testApp := test.NewApp()
	ui := NewUI(engine, testApp)

	// 3. Trigger refresh (what the background loop usually does)
	ui.refresh()

	// 4. Assert UI labels updated correctly
	if ui.lblWorkers.Text != "7 Goroutines" {
		t.Errorf("Expected '7 Goroutines', got '%s'", ui.lblWorkers.Text)
	}
	if ui.lblBatch.Text != "3 URLs" {
		t.Errorf("Expected '3 URLs', got '%s'", ui.lblBatch.Text)
	}
	if ui.lblMode.Text != "TESTING_MODE" {
		t.Errorf("Expected 'TESTING_MODE', got '%s'", ui.lblMode.Text)
	}
	
	// Check Telemetry mapping
	if ui.lblTotalIngested.Text != "2.50 MB" {
		t.Errorf("Expected '2.50 MB', got '%s'", ui.lblTotalIngested.Text)
	}
	if ui.lblQueueDepth.Text != "3 URLs" {
		t.Errorf("Expected '3 URLs', got '%s'", ui.lblQueueDepth.Text)
	}
	if ui.lblSuccessCount.Text != "95" {
		t.Errorf("Expected '95', got '%s'", ui.lblSuccessCount.Text)
	}
	if ui.lblFailureCount.Text != "5" {
		t.Errorf("Expected '5', got '%s'", ui.lblFailureCount.Text)
	}

	// 95 successes out of 100 total = 0.95 reliability score
	if ui.progressReliability.Value != 0.95 {
		t.Errorf("Expected reliability 0.95, got %f", ui.progressReliability.Value)
	}
	
	// Check Strategy mapping
	if ui.lblRetries.Text != "1" { // Default from NewEngine
		t.Errorf("Expected '1' max retries, got '%s'", ui.lblRetries.Text)
	}
	if ui.lblBackoff.Text != "1.50x" { // Default from NewEngine
		t.Errorf("Expected '1.50x' backoff, got '%s'", ui.lblBackoff.Text)
	}
}