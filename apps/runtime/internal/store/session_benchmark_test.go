package store

import (
	"os"
	"sync"
	"testing"
	"time"
)

// TestSessionPerformanceBasic tests basic session performance
func TestSessionPerformanceBasic(t *testing.T) {
	// Create temporary database
	tmpFile, err := os.CreateTemp("", "simple_perf_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	s, err := New(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer s.Close()

	// Test 1: Create 100 sessions
	start := time.Now()
	for i := 0; i < 100; i++ {
		_, err := s.CreateSession("Test Session")
		if err != nil {
			t.Errorf("Failed to create session: %v", err)
		}
	}
	createTime := time.Since(start)
	t.Logf("Created 100 sessions in %v (avg: %v)", createTime, createTime/100)

	// Test 2: List sessions (should be fast with indexes)
	start = time.Now()
	for i := 0; i < 100; i++ {
		sessions, err := s.ListSessions()
		if err != nil {
			t.Errorf("Failed to list sessions: %v", err)
		}
		if len(sessions) == 0 {
			t.Error("No sessions found")
			break
		}
	}
	listTime := time.Since(start)
	t.Logf("Listed sessions 100 times in %v (avg: %v)", listTime, listTime/100)

	// Test 3: Concurrent access
	var wg sync.WaitGroup
	const numWorkers = 50
	const opsPerWorker = 20

	start = time.Now()
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < opsPerWorker; j++ {
				// Mix of read and write operations
				sessions, _ := s.ListSessions()
				if workerID%5 == 0 && j%5 == 0 {
					s.CreateSession("Concurrent Session")
				}
				_ = len(sessions)
			}
		}(i)
	}
	wg.Wait()
	concurrentTime := time.Since(start)
	totalOps := numWorkers * opsPerWorker
	t.Logf("Concurrent access: %d ops in %v (avg: %v)", totalOps, concurrentTime, concurrentTime/time.Duration(totalOps))
}
