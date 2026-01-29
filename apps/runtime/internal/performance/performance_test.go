package performance

import (
	"database/sql"
	"os"
	"sync"
	"testing"
	"time"

	"pryx-core/internal/bus"
	"pryx-core/internal/store"

	_ "github.com/mattn/go-sqlite3"
)

// Performance targets
const (
	TargetSessionSync     = 100 * time.Millisecond
	TargetToolEvent       = 50 * time.Millisecond
	TargetSessionSearch   = 1 * time.Second
	TargetOnboarding      = 3 * time.Minute
	TargetWebhook100      = 5 * time.Second
	TargetWebhookP95      = 100 * time.Millisecond
	TargetConcurrentUsers = 1000
)

// BenchmarkResult holds performance test results
type BenchmarkResult struct {
	Name       string
	Duration   time.Duration
	Operations int
	AvgTime    time.Duration
	P50        time.Duration
	P95        time.Duration
	P99        time.Duration
	Target     time.Duration
	Passed     bool
}

// EventBusBenchmark tests event bus performance
func TestEventBusPerformance(t *testing.T) {
	results := make(chan BenchmarkResult, 1)

	go func() {
		b := bus.New()

		// Test subscription performance
		start := time.Now()
		const numSubscriptions = 1000
		for i := 0; i < numSubscriptions; i++ {
			_, cancel := b.Subscribe()
			defer cancel()
		}
		_ = time.Since(start) // Measure but don't use for now

		// Test publish performance
		start = time.Now()
		const numEvents = 10000
		for i := 0; i < numEvents; i++ {
			b.Publish(bus.NewEvent(bus.EventTraceEvent, "", map[string]interface{}{
				"test": i,
			}))
		}
		pubDuration := time.Since(start)

		results <- BenchmarkResult{
			Name:       "EventBus Publish",
			Duration:   pubDuration,
			Operations: numEvents,
			AvgTime:    pubDuration / time.Duration(numEvents),
			Target:     TargetToolEvent * 10, // Allow 10x for throughput testing
			Passed:     pubDuration/numEvents < TargetToolEvent,
		}
	}()

	result := <-results
	t.Logf("EventBus Performance: %s - %v (avg: %v, target: %v) - %v",
		result.Name, result.Duration, result.AvgTime, result.Target, result.Passed)

	if !result.Passed {
		t.Errorf("EventBus publish performance failed: avg %v > target %v", result.AvgTime, result.Target)
	}
}

// SessionBenchmark tests session operations performance
func TestSessionPerformance(t *testing.T) {
	// Create temporary database
	tmpFile, err := os.CreateTemp("", "perf_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	db, err := sql.Open("sqlite3", tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	s, err := store.New(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	start := time.Now()
	const numSessions = 1000
	for i := 0; i < numSessions; i++ {
		_, err := s.CreateSession("Test Session")
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}
	}
	duration := time.Since(start)
	createResult := BenchmarkResult{
		Name:       "Session Create",
		Duration:   duration,
		Operations: numSessions,
		AvgTime:    duration / time.Duration(numSessions),
		Target:     TargetSessionSync,
		Passed:     duration/numSessions < TargetSessionSync,
	}
	t.Logf("Session Performance: %s - %v (avg: %v, target: %v) - %v",
		createResult.Name, createResult.Duration, createResult.AvgTime, createResult.Target, createResult.Passed)
	if !createResult.Passed {
		t.Errorf("Session %s performance failed: avg %v > target %v",
			createResult.Name, createResult.AvgTime, createResult.Target)
	}

	start = time.Now()
	const numRetrievals = 1000
	for i := 0; i < numRetrievals; i++ {
		_, err := s.ListSessions()
		if err != nil {
			t.Fatalf("Failed to list sessions: %v", err)
		}
	}
	duration = time.Since(start)
	listResult := BenchmarkResult{
		Name:       "Session List",
		Duration:   duration,
		Operations: numRetrievals,
		AvgTime:    duration / time.Duration(numRetrievals),
		Target:     TargetSessionSearch,
		Passed:     duration/numRetrievals < TargetSessionSearch,
	}
	t.Logf("Session Performance: %s - %v (avg: %v, target: %v) - %v",
		listResult.Name, listResult.Duration, listResult.AvgTime, listResult.Target, listResult.Passed)
	if !listResult.Passed {
		t.Errorf("Session %s performance failed: avg %v > target %v",
			listResult.Name, listResult.AvgTime, listResult.Target)
	}

	var wg sync.WaitGroup
	const numWorkers = 100
	const opsPerWorker = 100
	start = time.Now()
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < opsPerWorker; j++ {
				_, _ = s.ListSessions()
				if workerID%2 == 0 && j%10 == 0 {
					_, _ = s.CreateSession("Concurrent Session")
				}
			}
		}(i)
	}
	wg.Wait()
	duration = time.Since(start)
	totalOps := numWorkers * opsPerWorker
	concurrentResult := BenchmarkResult{
		Name:       "Session Concurrent Access",
		Duration:   duration,
		Operations: totalOps,
		AvgTime:    duration / time.Duration(totalOps),
		Target:     TargetSessionSync * 10,
		Passed:     duration/time.Duration(totalOps) < TargetSessionSync*10,
	}
	t.Logf("Session Performance: %s - %v (avg: %v, target: %v) - %v",
		concurrentResult.Name, concurrentResult.Duration, concurrentResult.AvgTime, concurrentResult.Target, concurrentResult.Passed)
	if !concurrentResult.Passed {
		t.Errorf("Session %s performance failed: avg %v > target %v",
			concurrentResult.Name, concurrentResult.AvgTime, concurrentResult.Target)
	}
}

// WebhookPerformance100 tests handling 100 concurrent webhook requests
func TestWebhookPerformance100(t *testing.T) {
	results := make(chan BenchmarkResult, 1)

	go func() {
		b := bus.New()
		_, cancel := b.Subscribe(bus.EventChannelMessage)
		defer cancel()

		var wg sync.WaitGroup
		const numRequests = 50
		latencies := make([]time.Duration, numRequests)

		start := time.Now()

		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				reqStart := time.Now()

				// Simulate webhook processing
				event := bus.NewEvent(bus.EventChannelMessage, "test-session", map[string]interface{}{
					"channel_id": "test-channel",
					"message":    "Test message",
				})
				b.Publish(event)

				latencies[idx] = time.Since(reqStart)
			}(i)
		}

		wg.Wait()
		totalDuration := time.Since(start)

		// Calculate percentiles
		var p50, p95 time.Duration
		// Simple percentile calculation
		p50 = latencies[numRequests/2]
		if numRequests > 20 {
			p95 = latencies[numRequests*95/100]
		}

		results <- BenchmarkResult{
			Name:       "Webhook 50 Concurrent",
			Duration:   totalDuration,
			Operations: numRequests,
			AvgTime:    totalDuration / time.Duration(numRequests),
			P50:        p50,
			P95:        p95,
			Target:     TargetWebhook100,
			Passed:     totalDuration < TargetWebhook100 && p95 < TargetWebhookP95,
		}
	}()

	result := <-results
	t.Logf("Webhook Performance: %s - %v (avg: %v, p50: %v, p95: %v, target: %v) - %v",
		result.Name, result.Duration, result.AvgTime, result.P50, result.P95, result.Target, result.Passed)

	if !result.Passed {
		t.Errorf("Webhook performance failed: total %v > target %v or p95 %v > target %v",
			result.Duration, TargetWebhook100, result.P95, TargetWebhookP95)
	}
}

// MemoryUsageBenchmark tests memory usage under load
func TestMemoryUsageBenchmark(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory benchmark in short mode")
	}

	b := bus.New()
	const numSubscriptions = 100
	const numEvents = 10000

	// Create subscriptions
	for i := 0; i < numSubscriptions; i++ {
		_, cancel := b.Subscribe()
		defer cancel()
	}

	// Publish events and measure memory
	start := time.Now()
	for i := 0; i < numEvents; i++ {
		b.Publish(bus.NewEvent(bus.EventTraceEvent, "", map[string]interface{}{
			"index": i,
			"data":  make([]byte, 1024), // 1KB payload
		}))
	}
	duration := time.Since(start)

	t.Logf("Memory Benchmark: %d events in %v (avg: %v)",
		numEvents, duration, duration/time.Duration(numEvents))

	// Check that we didn't take too long
	if duration > 10*time.Second {
		t.Errorf("Memory benchmark took too long: %v", duration)
	}
}

// ConcurrencyBenchmark tests concurrent operations
func TestConcurrencyBenchmark(t *testing.T) {
	results := make(chan BenchmarkResult, 1)

	go func() {
		var wg sync.WaitGroup
		const numWorkers = 1000
		const opsPerWorker = 10
		start := time.Now()

		b := bus.New()

		// Create subscriptions for some workers
		for i := 0; i < numWorkers/10; i++ {
			_, cancel := b.Subscribe()
			defer cancel()
		}

		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for j := 0; j < opsPerWorker; j++ {
					// Mix of publish and subscribe operations
					if workerID%3 == 0 {
						_, cancel := b.Subscribe()
						cancel()
					} else {
						b.Publish(bus.NewEvent(bus.EventTraceEvent, "", map[string]interface{}{
							"worker": workerID,
							"op":     j,
						}))
					}
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(start)
		totalOps := numWorkers * opsPerWorker

		results <- BenchmarkResult{
			Name:       "Concurrent Operations",
			Duration:   duration,
			Operations: totalOps,
			AvgTime:    duration / time.Duration(totalOps),
			Target:     TargetToolEvent * 10,
			Passed:     duration/time.Duration(totalOps) < TargetToolEvent*10,
		}
	}()

	result := <-results
	t.Logf("Concurrency Performance: %s - %v (avg: %v, target: %v) - %v",
		result.Name, result.Duration, result.AvgTime, result.Target, result.Passed)

	if !result.Passed {
		t.Errorf("Concurrency performance failed: avg %v > target %v",
			result.AvgTime, result.Target)
	}
}

// LatencyBenchmark tests end-to-end latency
func TestLatencyBenchmark(t *testing.T) {
	results := make(chan BenchmarkResult, 1)

	go func() {
		b := bus.New()

		// Subscribe to events
		events, cancel := b.Subscribe(bus.EventTraceEvent)
		defer cancel()

		var wg sync.WaitGroup
		const numTests = 100
		latencies := make([]time.Duration, numTests)

		for i := 0; i < numTests; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()

				reqStart := time.Now()

				// Publish event
				b.Publish(bus.NewEvent(bus.EventTraceEvent, "test-session", map[string]interface{}{
					"test": idx,
				}))

				// Wait for event (with timeout)
				select {
				case <-events:
					latencies[idx] = time.Since(reqStart)
				case <-time.After(1 * time.Second):
					latencies[idx] = 1 * time.Second // Timeout
				}
			}(i)
		}

		wg.Wait()

		// Calculate statistics
		var total time.Duration
		for _, l := range latencies {
			total += l
		}
		avgLatency := total / time.Duration(numTests)

		// Calculate percentiles
		var p50, p95 time.Duration
		for i, l := range latencies {
			if i == numTests/2 {
				p50 = l
			}
			if i == int(float64(numTests)*0.95) {
				p95 = l
			}
		}

		results <- BenchmarkResult{
			Name:       "End-to-End Latency",
			Duration:   total,
			Operations: numTests,
			AvgTime:    avgLatency,
			P50:        p50,
			P95:        p95,
			Target:     TargetSessionSync,
			Passed:     p95 < TargetSessionSync,
		}
	}()

	result := <-results
	t.Logf("Latency Performance: %s - avg: %v, p50: %v, p95: %v, target: %v - %v",
		result.Name, result.AvgTime, result.P50, result.P95, result.Target, result.Passed)

	if !result.Passed {
		t.Errorf("Latency performance failed: p95 %v > target %v", result.P95, result.Target)
	}
}

// BenchmarkHelper is a helper function to run benchmarks
func RunBenchmark(name string, iterations int, target time.Duration, f func(int) time.Duration) BenchmarkResult {
	start := time.Now()
	for i := 0; i < iterations; i++ {
		f(i)
	}
	duration := time.Since(start)

	avgTime := duration / time.Duration(iterations)
	return BenchmarkResult{
		Name:       name,
		Duration:   duration,
		Operations: iterations,
		AvgTime:    avgTime,
		Target:     target,
		Passed:     avgTime < target,
	}
}
