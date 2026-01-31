// Package performance provides memory profiling and optimization utilities.
package performance

import (
	"fmt"
	"log"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// ComponentMemory tracks memory usage for a specific component
type ComponentMemory struct {
	Name        string
	AllocBytes  uint64
	TotalBytes  uint64
	ObjectCount int
	LastUpdated time.Time
}

// MemorySnapshot captures memory stats at a point in time
type MemorySnapshot struct {
	Timestamp    time.Time
	AllocBytes   uint64
	TotalBytes   uint64
	SysBytes     uint64
	NumGC        uint32
	HeapAlloc    uint64
	HeapSys      uint64
	HeapIdle     uint64
	HeapInuse    uint64
	HeapReleased uint64
	HeapObjects  uint64
}

// MemoryLimit defines memory constraints
type MemoryLimit struct {
	MaxAllocBytes   uint64  // Max allocated memory before warning
	MaxSysBytes     uint64  // Max system memory before critical
	WarningPercent  float64 // Percentage at which to warn
	CriticalPercent float64 // Percentage at which to take action
}

// Default memory limits (100MB base target)
var DefaultMemoryLimits = MemoryLimit{
	MaxAllocBytes:   100 * 1024 * 1024, // 100MB
	MaxSysBytes:     150 * 1024 * 1024, // 150MB
	WarningPercent:  0.80,              // 80%
	CriticalPercent: 0.95,              // 95%
}

// MemoryProfiler tracks memory usage across components
type MemoryProfiler struct {
	mu            sync.RWMutex
	startTime     time.Time
	components    map[string]*ComponentMemory
	snapshots     []MemorySnapshot
	limits        MemoryLimit
	logger        *log.Logger
	enabled       bool
	lastSnapshot  time.Time
	snapshotEvery time.Duration
	onWarning     func(usage MemorySnapshot, limit MemoryLimit)
	onCritical    func(usage MemorySnapshot, limit MemoryLimit)
}

// NewMemoryProfiler creates a new memory profiler
func NewMemoryProfiler() *MemoryProfiler {
	return &MemoryProfiler{
		startTime:     time.Now(),
		components:    make(map[string]*ComponentMemory),
		snapshots:     make([]MemorySnapshot, 0),
		limits:        DefaultMemoryLimits,
		logger:        log.New(log.Writer(), "[MEMORY] ", log.LstdFlags|log.Lmicroseconds),
		enabled:       true,
		snapshotEvery: 30 * time.Second,
	}
}

// NewMemoryProfilerWithLogger creates a profiler with a custom logger
func NewMemoryProfilerWithLogger(logger *log.Logger) *MemoryProfiler {
	return &MemoryProfiler{
		startTime:     time.Now(),
		components:    make(map[string]*ComponentMemory),
		snapshots:     make([]MemorySnapshot, 0),
		limits:        DefaultMemoryLimits,
		logger:        logger,
		enabled:       true,
		snapshotEvery: 30 * time.Second,
	}
}

// SetLimits configures memory limits
func (mp *MemoryProfiler) SetLimits(limits MemoryLimit) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.limits = limits
}

// SetCallbacks sets warning and critical callbacks
func (mp *MemoryProfiler) SetCallbacks(onWarning, onCritical func(usage MemorySnapshot, limit MemoryLimit)) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.onWarning = onWarning
	mp.onCritical = onCritical
}

// RecordComponent records memory usage for a component
func (mp *MemoryProfiler) RecordComponent(name string, allocBytes uint64, objectCount int) {
	if !mp.enabled {
		return
	}

	mp.mu.Lock()
	defer mp.mu.Unlock()

	component, exists := mp.components[name]
	if !exists {
		component = &ComponentMemory{Name: name}
		mp.components[name] = component
	}

	component.AllocBytes = allocBytes
	component.TotalBytes += allocBytes
	component.ObjectCount = objectCount
	component.LastUpdated = time.Now()
}

// UpdateComponent updates component memory (adds delta)
func (mp *MemoryProfiler) UpdateComponent(name string, deltaBytes int64, deltaObjects int) {
	if !mp.enabled {
		return
	}

	mp.mu.Lock()
	defer mp.mu.Unlock()

	component, exists := mp.components[name]
	if !exists {
		component = &ComponentMemory{Name: name}
		mp.components[name] = component
	}

	if deltaBytes > 0 {
		component.AllocBytes += uint64(deltaBytes)
		component.TotalBytes += uint64(deltaBytes)
	} else if deltaBytes < 0 && component.AllocBytes >= uint64(-deltaBytes) {
		component.AllocBytes -= uint64(-deltaBytes)
	}

	component.ObjectCount += deltaObjects
	if component.ObjectCount < 0 {
		component.ObjectCount = 0
	}
	component.LastUpdated = time.Now()
}

// GetCurrentSnapshot captures current memory stats
func (mp *MemoryProfiler) GetCurrentSnapshot() MemorySnapshot {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemorySnapshot{
		Timestamp:    time.Now(),
		AllocBytes:   m.Alloc,
		TotalBytes:   m.TotalAlloc,
		SysBytes:     m.Sys,
		NumGC:        m.NumGC,
		HeapAlloc:    m.HeapAlloc,
		HeapSys:      m.HeapSys,
		HeapIdle:     m.HeapIdle,
		HeapInuse:    m.HeapInuse,
		HeapReleased: m.HeapReleased,
		HeapObjects:  m.HeapObjects,
	}
}

// TakeSnapshot records a memory snapshot
func (mp *MemoryProfiler) TakeSnapshot() MemorySnapshot {
	mp.mu.RLock()
	enabled := mp.enabled
	mp.mu.RUnlock()

	if !enabled {
		return MemorySnapshot{}
	}

	mp.mu.Lock()
	defer mp.mu.Unlock()

	snapshot := mp.GetCurrentSnapshot()
	mp.snapshots = append(mp.snapshots, snapshot)
	mp.lastSnapshot = snapshot.Timestamp

	// Keep only last 100 snapshots to prevent unbounded growth
	if len(mp.snapshots) > 100 {
		mp.snapshots = mp.snapshots[len(mp.snapshots)-100:]
	}

	// Check limits and trigger callbacks
	mp.checkLimits(snapshot)

	return snapshot
}

// checkLimits checks memory against limits and triggers callbacks
func (mp *MemoryProfiler) checkLimits(snapshot MemorySnapshot) {
	allocPercent := float64(snapshot.AllocBytes) / float64(mp.limits.MaxAllocBytes)
	sysPercent := float64(snapshot.SysBytes) / float64(mp.limits.MaxSysBytes)

	if sysPercent >= mp.limits.CriticalPercent || allocPercent >= mp.limits.CriticalPercent {
		mp.logger.Printf("⚠ CRITICAL: Memory usage at %.1f%% (alloc: %s, sys: %s)",
			max(allocPercent, sysPercent)*100,
			formatBytes(snapshot.AllocBytes),
			formatBytes(snapshot.SysBytes))
		if mp.onCritical != nil {
			go mp.onCritical(snapshot, mp.limits)
		}
	} else if sysPercent >= mp.limits.WarningPercent || allocPercent >= mp.limits.WarningPercent {
		mp.logger.Printf("⚡ WARNING: Memory usage at %.1f%% (alloc: %s, sys: %s)",
			max(allocPercent, sysPercent)*100,
			formatBytes(snapshot.AllocBytes),
			formatBytes(snapshot.SysBytes))
		if mp.onWarning != nil {
			go mp.onWarning(snapshot, mp.limits)
		}
	}
}

// StartMonitoring begins periodic memory monitoring
func (mp *MemoryProfiler) StartMonitoring() {
	mp.mu.RLock()
	enabled := mp.enabled
	snapshotEvery := mp.snapshotEvery
	mp.mu.RUnlock()

	if !enabled {
		return
	}

	go func() {
		ticker := time.NewTicker(snapshotEvery)
		defer ticker.Stop()

		for range ticker.C {
			mp.mu.RLock()
			shouldContinue := mp.enabled
			mp.mu.RUnlock()

			if !shouldContinue {
				return
			}
			mp.TakeSnapshot()
		}
	}()
}

// GetComponents returns all component memory stats
func (mp *MemoryProfiler) GetComponents() []ComponentMemory {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	components := make([]ComponentMemory, 0, len(mp.components))
	for _, c := range mp.components {
		components = append(components, *c)
	}

	// Sort by allocated bytes (descending)
	sort.Slice(components, func(i, j int) bool {
		return components[i].AllocBytes > components[j].AllocBytes
	})

	return components
}

// GetSnapshots returns all recorded snapshots
func (mp *MemoryProfiler) GetSnapshots() []MemorySnapshot {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	snapshots := make([]MemorySnapshot, len(mp.snapshots))
	copy(snapshots, mp.snapshots)
	return snapshots
}

// GetComponent returns a specific component by name
func (mp *MemoryProfiler) GetComponent(name string) (ComponentMemory, bool) {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	component, exists := mp.components[name]
	if !exists {
		return ComponentMemory{}, false
	}
	return *component, true
}

// GetTotalComponentMemory returns sum of all component allocations
func (mp *MemoryProfiler) GetTotalComponentMemory() uint64 {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	var total uint64
	for _, c := range mp.components {
		total += c.AllocBytes
	}
	return total
}

// Reset clears all component data
func (mp *MemoryProfiler) Reset() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.components = make(map[string]*ComponentMemory)
	mp.snapshots = make([]MemorySnapshot, 0)
	mp.startTime = time.Now()
}

// GenerateReport creates a formatted memory report
func (mp *MemoryProfiler) GenerateReport() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	current := mp.GetCurrentSnapshot()
	components := mp.GetComponents()
	uptime := time.Since(mp.startTime)

	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("╔════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║           MEMORY USAGE REPORT                          ║\n")
	sb.WriteString("╠════════════════════════════════════════════════════════╣\n")
	sb.WriteString(fmt.Sprintf("║ Uptime: %-46s ║\n", formatDuration(uptime)))
	sb.WriteString(fmt.Sprintf("║ Allocated: %-43s ║\n", formatBytes(current.AllocBytes)))
	sb.WriteString(fmt.Sprintf("║ System: %-46s ║\n", formatBytes(current.SysBytes)))
	sb.WriteString(fmt.Sprintf("║ Heap In-Use: %-41s ║\n", formatBytes(current.HeapInuse)))
	sb.WriteString(fmt.Sprintf("║ Heap Objects: %-40d ║\n", current.HeapObjects))
	sb.WriteString(fmt.Sprintf("║ GC Cycles: %-43d ║\n", current.NumGC))
	sb.WriteString("╠════════════════════════════════════════════════════════╣\n")
	sb.WriteString(fmt.Sprintf("║ %-20s %-12s %s ║\n", "Component", "Memory", "Objects"))
	sb.WriteString("╠════════════════════════════════════════════════════════╣\n")

	if len(components) == 0 {
		sb.WriteString(fmt.Sprintf("║ %-20s ║\n", "No component data recorded"))
	} else {
		for _, c := range components {
			sb.WriteString(fmt.Sprintf("║ %-20s %-12s %5d ║\n",
				truncateString(c.Name, 20),
				formatBytes(c.AllocBytes),
				c.ObjectCount))
		}
	}

	sb.WriteString("╚════════════════════════════════════════════════════════╝\n")

	return sb.String()
}

// PrintReport logs the memory report
func (mp *MemoryProfiler) PrintReport() {
	if !mp.enabled {
		return
	}
	mp.logger.Print(mp.GenerateReport())
}

// Enable enables profiling
func (mp *MemoryProfiler) Enable() {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.enabled = true
}

// Disable disables profiling
func (mp *MemoryProfiler) Disable() {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.enabled = false
}

// IsEnabled returns true if profiling is enabled
func (mp *MemoryProfiler) IsEnabled() bool {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.enabled
}

// ForceGC triggers garbage collection and returns stats
func (mp *MemoryProfiler) ForceGC() MemorySnapshot {
	runtime.GC()
	return mp.TakeSnapshot()
}

// GetMemoryStats returns current memory statistics
func GetMemoryStats() MemorySnapshot {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemorySnapshot{
		Timestamp:    time.Now(),
		AllocBytes:   m.Alloc,
		TotalBytes:   m.TotalAlloc,
		SysBytes:     m.Sys,
		NumGC:        m.NumGC,
		HeapAlloc:    m.HeapAlloc,
		HeapSys:      m.HeapSys,
		HeapIdle:     m.HeapIdle,
		HeapInuse:    m.HeapInuse,
		HeapReleased: m.HeapReleased,
		HeapObjects:  m.HeapObjects,
	}
}

// formatBytes formats bytes in human-readable format
func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// max returns the maximum of two float64 values
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
