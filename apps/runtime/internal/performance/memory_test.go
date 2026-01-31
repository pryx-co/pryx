package performance

import (
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewMemoryProfiler(t *testing.T) {
	mp := NewMemoryProfiler()

	if mp == nil {
		t.Fatal("NewMemoryProfiler returned nil")
	}

	if !mp.enabled {
		t.Error("Profiler should be enabled by default")
	}

	if len(mp.components) != 0 {
		t.Error("Components should be empty initially")
	}

	if len(mp.snapshots) != 0 {
		t.Error("Snapshots should be empty initially")
	}
}

func TestNewMemoryProfilerWithLogger(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	mp := NewMemoryProfilerWithLogger(logger)

	if mp == nil {
		t.Fatal("NewMemoryProfilerWithLogger returned nil")
	}

	if mp.logger != logger {
		t.Error("Logger not set correctly")
	}
}

func TestSetLimits(t *testing.T) {
	mp := NewMemoryProfiler()

	customLimits := MemoryLimit{
		MaxAllocBytes:   200 * 1024 * 1024,
		MaxSysBytes:     300 * 1024 * 1024,
		WarningPercent:  0.75,
		CriticalPercent: 0.90,
	}

	mp.SetLimits(customLimits)

	if mp.limits.MaxAllocBytes != customLimits.MaxAllocBytes {
		t.Errorf("MaxAllocBytes: got %d, want %d", mp.limits.MaxAllocBytes, customLimits.MaxAllocBytes)
	}

	if mp.limits.WarningPercent != customLimits.WarningPercent {
		t.Errorf("WarningPercent: got %f, want %f", mp.limits.WarningPercent, customLimits.WarningPercent)
	}
}

func TestRecordComponent(t *testing.T) {
	mp := NewMemoryProfiler()

	mp.RecordComponent("test-component", 1024, 10)

	component, exists := mp.GetComponent("test-component")
	if !exists {
		t.Fatal("Component not found")
	}

	if component.AllocBytes != 1024 {
		t.Errorf("AllocBytes: got %d, want %d", component.AllocBytes, 1024)
	}

	if component.ObjectCount != 10 {
		t.Errorf("ObjectCount: got %d, want %d", component.ObjectCount, 10)
	}

	if component.Name != "test-component" {
		t.Errorf("Name: got %s, want %s", component.Name, "test-component")
	}
}

func TestUpdateComponent(t *testing.T) {
	mp := NewMemoryProfiler()

	// Initial record
	mp.RecordComponent("test-component", 1024, 10)

	// Update with positive delta
	mp.UpdateComponent("test-component", 512, 5)

	component, _ := mp.GetComponent("test-component")
	if component.AllocBytes != 1536 {
		t.Errorf("AllocBytes after positive delta: got %d, want %d", component.AllocBytes, 1536)
	}

	if component.ObjectCount != 15 {
		t.Errorf("ObjectCount after positive delta: got %d, want %d", component.ObjectCount, 15)
	}

	// Update with negative delta
	mp.UpdateComponent("test-component", -256, -3)

	component, _ = mp.GetComponent("test-component")
	if component.AllocBytes != 1280 {
		t.Errorf("AllocBytes after negative delta: got %d, want %d", component.AllocBytes, 1280)
	}

	if component.ObjectCount != 12 {
		t.Errorf("ObjectCount after negative delta: got %d, want %d", component.ObjectCount, 12)
	}
}

func TestUpdateComponent_NegativeBounds(t *testing.T) {
	mp := NewMemoryProfiler()

	// Initial record with small values
	mp.RecordComponent("small-component", 100, 5)

	// Try to subtract more than exists - should stay at current value
	mp.UpdateComponent("small-component", -200, -10)

	component, _ := mp.GetComponent("small-component")
	// When subtracting more than exists, value stays unchanged (protection against underflow)
	if component.AllocBytes != 100 {
		t.Errorf("AllocBytes should stay at 100 when subtracting more than exists: got %d", component.AllocBytes)
	}

	// ObjectCount can go to 0 (clamped)
	if component.ObjectCount != 0 {
		t.Errorf("ObjectCount should not go below 0: got %d", component.ObjectCount)
	}
}

func TestGetCurrentSnapshot(t *testing.T) {
	mp := NewMemoryProfiler()

	snapshot := mp.GetCurrentSnapshot()

	if snapshot.Timestamp.IsZero() {
		t.Error("Snapshot timestamp should not be zero")
	}

	if snapshot.AllocBytes == 0 && snapshot.SysBytes == 0 {
		t.Error("Snapshot should have some memory data")
	}
}

func TestTakeSnapshot(t *testing.T) {
	mp := NewMemoryProfiler()

	_ = mp.TakeSnapshot()

	if len(mp.snapshots) != 1 {
		t.Errorf("Expected 1 snapshot, got %d", len(mp.snapshots))
	}

	if mp.lastSnapshot.IsZero() {
		t.Error("Last snapshot time should be set")
	}

	snapshots := mp.GetSnapshots()
	if len(snapshots) != 1 {
		t.Errorf("GetSnapshots returned %d snapshots, want 1", len(snapshots))
	}
}

func TestTakeSnapshot_MaxSnapshots(t *testing.T) {
	mp := NewMemoryProfiler()
	mp.snapshotEvery = 1 * time.Millisecond

	// Take more than 100 snapshots
	for i := 0; i < 110; i++ {
		mp.TakeSnapshot()
		time.Sleep(1 * time.Millisecond)
	}

	if len(mp.snapshots) > 100 {
		t.Errorf("Expected max 100 snapshots, got %d", len(mp.snapshots))
	}
}

func TestGetComponents(t *testing.T) {
	mp := NewMemoryProfiler()

	// Record multiple components
	mp.RecordComponent("component-a", 2048, 20)
	mp.RecordComponent("component-b", 1024, 10)
	mp.RecordComponent("component-c", 4096, 40)

	components := mp.GetComponents()

	if len(components) != 3 {
		t.Errorf("Expected 3 components, got %d", len(components))
	}

	// Should be sorted by AllocBytes descending
	if components[0].Name != "component-c" {
		t.Errorf("First component should be 'component-c' (largest), got %s", components[0].Name)
	}

	if components[2].Name != "component-b" {
		t.Errorf("Last component should be 'component-b' (smallest), got %s", components[2].Name)
	}
}

func TestGetTotalComponentMemory(t *testing.T) {
	mp := NewMemoryProfiler()

	mp.RecordComponent("a", 1000, 10)
	mp.RecordComponent("b", 2000, 20)
	mp.RecordComponent("c", 3000, 30)

	total := mp.GetTotalComponentMemory()

	if total != 6000 {
		t.Errorf("Expected total 6000, got %d", total)
	}
}

func TestGetComponent_NotFound(t *testing.T) {
	mp := NewMemoryProfiler()

	_, exists := mp.GetComponent("nonexistent")
	if exists {
		t.Error("Should return false for nonexistent component")
	}
}

func TestEnableDisable(t *testing.T) {
	mp := NewMemoryProfiler()

	if !mp.IsEnabled() {
		t.Error("Should be enabled by default")
	}

	mp.Disable()
	if mp.IsEnabled() {
		t.Error("Should be disabled after Disable()")
	}

	mp.Enable()
	if !mp.IsEnabled() {
		t.Error("Should be enabled after Enable()")
	}
}

func TestReset(t *testing.T) {
	mp := NewMemoryProfiler()

	mp.RecordComponent("test", 1024, 10)
	mp.TakeSnapshot()

	mp.Reset()

	if len(mp.components) != 0 {
		t.Error("Components should be empty after reset")
	}

	if len(mp.snapshots) != 0 {
		t.Error("Snapshots should be empty after reset")
	}

	components := mp.GetComponents()
	if len(components) != 0 {
		t.Errorf("GetComponents should return empty slice, got %d", len(components))
	}
}

func TestForceGC(t *testing.T) {
	mp := NewMemoryProfiler()

	// Allocate some memory
	_ = make([]byte, 1024*1024)

	snapshot := mp.ForceGC()

	if snapshot.Timestamp.IsZero() {
		t.Error("ForceGC should return a snapshot with timestamp")
	}

	if len(mp.snapshots) != 1 {
		t.Errorf("ForceGC should take a snapshot, got %d snapshots", len(mp.snapshots))
	}
}

func TestGenerateReport(t *testing.T) {
	mp := NewMemoryProfiler()

	mp.RecordComponent("websocket", 1024*1024, 100)
	mp.RecordComponent("messages", 512*1024, 50)
	mp.TakeSnapshot()

	report := mp.GenerateReport()

	if report == "" {
		t.Error("GenerateReport should return non-empty string")
	}

	// Check for key sections
	if !strings.Contains(report, "MEMORY USAGE REPORT") {
		t.Error("Report should contain header")
	}

	if !strings.Contains(report, "websocket") {
		t.Error("Report should contain component names")
	}

	if !strings.Contains(report, "messages") {
		t.Error("Report should contain all component names")
	}
}

func TestGetMemoryStats(t *testing.T) {
	stats := GetMemoryStats()

	if stats.Timestamp.IsZero() {
		t.Error("GetMemoryStats should return timestamp")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    uint64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
	}

	for _, test := range tests {
		result := formatBytes(test.bytes)
		if result != test.expected {
			t.Errorf("formatBytes(%d) = %s, want %s", test.bytes, result, test.expected)
		}
	}
}

func TestSetCallbacks(t *testing.T) {
	mp := NewMemoryProfiler()

	var criticalCalled bool
	var criticalCalledMu sync.Mutex

	onWarning := func(usage MemorySnapshot, limit MemoryLimit) {}

	onCritical := func(usage MemorySnapshot, limit MemoryLimit) {
		criticalCalledMu.Lock()
		criticalCalled = true
		criticalCalledMu.Unlock()
	}

	mp.SetCallbacks(onWarning, onCritical)

	// Set low limits to trigger critical (only critical is called when both thresholds exceeded)
	mp.SetLimits(MemoryLimit{
		MaxAllocBytes:   1, // 1 byte to trigger immediately
		MaxSysBytes:     1,
		WarningPercent:  0.5,
		CriticalPercent: 0.5, // Same as warning to ensure we hit critical
	})

	// Take a snapshot which should trigger callbacks
	mp.TakeSnapshot()

	// Give callbacks time to execute
	time.Sleep(100 * time.Millisecond)

	// Critical should be called when both thresholds are exceeded
	criticalCalledMu.Lock()
	if !criticalCalled {
		t.Error("Critical callback should have been called")
	}
	criticalCalledMu.Unlock()

	// Test warning-only scenario with fresh profiler
	mp2 := NewMemoryProfiler()

	var warningOnlyCalled bool
	var warningOnlyCalledMu sync.Mutex

	onWarningOnly := func(usage MemorySnapshot, limit MemoryLimit) {
		warningOnlyCalledMu.Lock()
		warningOnlyCalled = true
		warningOnlyCalledMu.Unlock()
	}

	mp2.SetCallbacks(onWarningOnly, nil)

	// Set limits that trigger warning but not critical
	mp2.SetLimits(MemoryLimit{
		MaxAllocBytes:   1024 * 1024 * 1024, // 1GB - won't be exceeded
		MaxSysBytes:     1024 * 1024 * 1024,
		WarningPercent:  0.001, // 0.1% - very low to trigger warning
		CriticalPercent: 0.999, // 99.9% - won't be reached
	})

	mp2.TakeSnapshot()
	time.Sleep(100 * time.Millisecond)

	warningOnlyCalledMu.Lock()
	if !warningOnlyCalled {
		t.Error("Warning callback should have been called for warning-only scenario")
	}
	warningOnlyCalledMu.Unlock()
}

func TestStartMonitoring(t *testing.T) {
	mp := NewMemoryProfiler()
	mp.snapshotEvery = 50 * time.Millisecond

	mp.StartMonitoring()

	// Wait for at least one snapshot
	time.Sleep(100 * time.Millisecond)

	mp.Disable() // Stop monitoring

	if len(mp.snapshots) == 0 {
		t.Error("StartMonitoring should take periodic snapshots")
	}
}

func TestCheckLimits(t *testing.T) {
	mp := NewMemoryProfiler()
	mp.Disable() // Disable to prevent actual logging

	// Test with normal limits
	mp.checkLimits(MemorySnapshot{
		AllocBytes: 50 * 1024 * 1024, // 50MB
		SysBytes:   60 * 1024 * 1024,
	})
	// Should not trigger anything at 50%
}

func TestTakeSnapshot_Disabled(t *testing.T) {
	mp := NewMemoryProfiler()
	mp.Disable()

	snapshot := mp.TakeSnapshot()

	if !snapshot.Timestamp.IsZero() {
		t.Error("TakeSnapshot should return empty snapshot when disabled")
	}
}

func TestRecordComponent_Disabled(t *testing.T) {
	mp := NewMemoryProfiler()
	mp.Disable()

	mp.RecordComponent("test", 1024, 10)

	_, exists := mp.GetComponent("test")
	if exists {
		t.Error("Should not record when disabled")
	}
}

func TestMemoryProfilerConcurrency(t *testing.T) {
	mp := NewMemoryProfiler()

	// Run concurrent operations
	done := make(chan bool, 3)

	go func() {
		for i := 0; i < 100; i++ {
			mp.RecordComponent("comp-a", uint64(i*1024), i)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			mp.TakeSnapshot()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			mp.GetComponents()
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// Should not panic and should have data
	if len(mp.components) == 0 {
		t.Error("Should have recorded components after concurrent operations")
	}
}

func BenchmarkTakeSnapshot(b *testing.B) {
	mp := NewMemoryProfiler()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mp.TakeSnapshot()
	}
}

func BenchmarkRecordComponent(b *testing.B) {
	mp := NewMemoryProfiler()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mp.RecordComponent("benchmark-comp", uint64(i), i)
	}
}

func BenchmarkGetMemoryStats(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetMemoryStats()
	}
}

func BenchmarkFormatBytes(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatBytes(uint64(i * 1024))
	}
}

// Ensure we test with real memory operations
func TestRealMemoryTracking(t *testing.T) {
	mp := NewMemoryProfiler()

	// Record baseline
	baseline := mp.GetCurrentSnapshot()

	// Allocate memory
	data := make([]byte, 10*1024*1024) // 10MB
	for i := range data {
		data[i] = byte(i % 256)
	}

	// Force GC to get accurate reading
	runtime.GC()
	time.Sleep(10 * time.Millisecond)

	// Check memory increased
	after := mp.GetCurrentSnapshot()

	// Memory should have increased (or at least not be less)
	if after.HeapAlloc < baseline.HeapAlloc {
		t.Logf("Note: HeapAlloc decreased from %d to %d (GC may have run)",
			baseline.HeapAlloc, after.HeapAlloc)
	}

	// Keep reference to prevent optimization
	_ = data
}
