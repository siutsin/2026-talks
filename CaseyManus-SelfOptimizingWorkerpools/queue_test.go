package main

import (
	"testing"
)

func TestQueuePushAndLen(t *testing.T) {
	q := NewURLQueue()
	if q.Len() != 0 {
		t.Errorf("Expected queue length to be 0, got %d", q.Len())
	}

	urls := []string{"http://example.com/1", "http://example.com/2"}
	q.Push(urls)
	
	if q.Len() != 2 {
		t.Errorf("Expected queue length to be 2, got %d", q.Len())
	}
}

func TestQueuePushDeduplication(t *testing.T) {
	q := NewURLQueue()
	urls := []string{"http://example.com/1", "http://example.com/1"}
	q.Push(urls)
	
	if q.Len() != 1 {
		t.Errorf("Expected deduplication to result in 1 item, got %d", q.Len())
	}
}

func TestQueuePopBatch(t *testing.T) {
	q := NewURLQueue()
	urls := []string{"http://example.com/1", "http://example.com/2", "http://example.com/3"}
	q.Push(urls)

	batch := q.PopBatch(2)
	if len(batch) != 2 {
		t.Errorf("Expected batch size of 2, got %d", len(batch))
	}
	if q.Len() != 1 {
		t.Errorf("Expected remaining queue length to be 1, got %d", q.Len())
	}

	batch2 := q.PopBatch(5) // Requesting more than available
	if len(batch2) != 1 {
		t.Errorf("Expected batch size of 1, got %d", len(batch2))
	}
	if q.Len() != 0 {
		t.Errorf("Expected queue to be empty, got %d", q.Len())
	}

	batch3 := q.PopBatch(1)
	if batch3 != nil {
		t.Errorf("Expected nil batch from empty queue, got %v", batch3)
	}
}