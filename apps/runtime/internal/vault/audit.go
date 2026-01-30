package vault

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	defaultAuditDir      = "~/.pryx/audit"
	defaultRetentionDays = 90
	maxLogFileSize       = 100 * 1024 * 1024 // 100MB
)

// AuditAction represents the type of vault operation
type AuditAction string

const (
	ActionUnlock AuditAction = "unlock"
	ActionLock   AuditAction = "lock"
	ActionRead   AuditAction = "read"
	ActionWrite  AuditAction = "write"
	ActionDelete AuditAction = "delete"
	ActionRotate AuditAction = "rotate"
	ActionList   AuditAction = "list"
	ActionExport AuditAction = "export"
	ActionImport AuditAction = "import"
)

// AuditEntry represents a single audit log entry
type AuditEntry struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Action    AuditAction            `json:"action"`
	Target    string                 `json:"target"`
	Actor     string                 `json:"actor"`
	Success   bool                   `json:"success"`
	Error     string                 `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	PrevHash  string                 `json:"prev_hash"`
	Hash      string                 `json:"hash"`
}

// AuditLogger handles vault audit logging with tamper detection
type AuditLogger struct {
	mu sync.RWMutex

	auditDir      string
	retentionDays int
	currentFile   *os.File
	currentDate   string
	lastHash      string

	buffer      []AuditEntry
	bufferSize  int
	flushTicker *time.Ticker
	done        chan struct{}
}

// AuditLoggerOption configures the audit logger
type AuditLoggerOption func(*AuditLogger)

// WithAuditDir sets the audit log directory
func WithAuditDir(dir string) AuditLoggerOption {
	return func(a *AuditLogger) {
		a.auditDir = dir
	}
}

// WithRetentionDays sets the log retention period
func WithRetentionDays(days int) AuditLoggerOption {
	return func(a *AuditLogger) {
		a.retentionDays = days
	}
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(opts ...AuditLoggerOption) (*AuditLogger, error) {
	a := &AuditLogger{
		auditDir:      defaultAuditDir,
		retentionDays: defaultRetentionDays,
		buffer:        make([]AuditEntry, 0, 100),
		done:          make(chan struct{}),
	}

	for _, opt := range opts {
		opt(a)
	}

	// Expand home directory
	if strings.HasPrefix(a.auditDir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		a.auditDir = filepath.Join(home, a.auditDir[2:])
	}

	// Create audit directory
	if err := os.MkdirAll(a.auditDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create audit directory: %w", err)
	}

	// Initialize current log file
	if err := a.rotateLogFile(); err != nil {
		return nil, fmt.Errorf("failed to initialize log file: %w", err)
	}

	// Start background flush ticker
	a.flushTicker = time.NewTicker(5 * time.Second)
	go a.backgroundFlush()

	return a, nil
}

// Log records an audit entry
func (a *AuditLogger) Log(action AuditAction, target, actor string, success bool, err error, metadata map[string]interface{}) error {
	entry := AuditEntry{
		ID:        uuid.New().String(),
		Timestamp: time.Now().UTC(),
		Action:    action,
		Target:    target,
		Actor:     actor,
		Success:   success,
		Metadata:  metadata,
		PrevHash:  a.lastHash,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	// Calculate hash for tamper detection
	entry.Hash = a.calculateHash(entry)

	a.mu.Lock()
	defer a.mu.Unlock()

	a.lastHash = entry.Hash
	a.buffer = append(a.buffer, entry)
	a.bufferSize++

	// Flush if buffer is full
	if a.bufferSize >= 100 {
		return a.flushBuffer()
	}

	return nil
}

// LogUnlock records a vault unlock operation
func (a *AuditLogger) LogUnlock(actor string, success bool, err error) error {
	return a.Log(ActionUnlock, "vault", actor, success, err, nil)
}

// LogLock records a vault lock operation
func (a *AuditLogger) LogLock(actor string) error {
	return a.Log(ActionLock, "vault", actor, true, nil, nil)
}

// LogRead records a credential read operation
func (a *AuditLogger) LogRead(credentialID, actor string, success bool, err error) error {
	return a.Log(ActionRead, credentialID, actor, success, err, nil)
}

// LogWrite records a credential write operation
func (a *AuditLogger) LogWrite(credentialID, actor string, success bool, err error) error {
	return a.Log(ActionWrite, credentialID, actor, success, err, nil)
}

// LogDelete records a credential delete operation
func (a *AuditLogger) LogDelete(credentialID, actor string, success bool, err error) error {
	return a.Log(ActionDelete, credentialID, actor, success, err, nil)
}

// LogRotate records a key rotation operation
func (a *AuditLogger) LogRotate(actor string, success bool, err error) error {
	return a.Log(ActionRotate, "vault", actor, success, err, nil)
}

// Query searches audit logs with filters
func (a *AuditLogger) Query(opts QueryOptions) ([]AuditEntry, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Flush any pending entries
	if err := a.flushBuffer(); err != nil {
		return nil, fmt.Errorf("failed to flush buffer: %w", err)
	}

	// Get list of log files to search
	files, err := a.getLogFiles(opts.StartTime, opts.EndTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get log files: %w", err)
	}

	var results []AuditEntry

	for _, file := range files {
		entries, err := a.readLogFile(file, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to read log file %s: %w", file, err)
		}
		results = append(results, entries...)
	}

	// Sort by timestamp
	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp.Before(results[j].Timestamp)
	})

	// Apply limit
	if opts.Limit > 0 && len(results) > opts.Limit {
		results = results[len(results)-opts.Limit:]
	}

	return results, nil
}

// QueryOptions defines filters for audit log queries
type QueryOptions struct {
	StartTime *time.Time
	EndTime   *time.Time
	Actions   []AuditAction
	Actor     string
	Target    string
	Success   *bool
	Limit     int
}

// Export exports audit logs to a file
func (a *AuditLogger) Export(startTime, endTime time.Time, format string, w *os.File) error {
	entries, err := a.Query(QueryOptions{
		StartTime: &startTime,
		EndTime:   &endTime,
	})
	if err != nil {
		return err
	}

	switch format {
	case "json":
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(entries)
	case "csv":
		return a.exportCSV(entries, w)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// VerifyIntegrity checks the integrity of audit logs using hash chain
func (a *AuditLogger) VerifyIntegrity(startTime, endTime time.Time) error {
	entries, err := a.Query(QueryOptions{
		StartTime: &startTime,
		EndTime:   &endTime,
	})
	if err != nil {
		return err
	}

	var prevHash string
	for i, entry := range entries {
		// Verify prev_hash matches
		if i > 0 && entry.PrevHash != prevHash {
			return fmt.Errorf("hash chain broken at entry %s: prev_hash mismatch", entry.ID)
		}

		// Verify entry hash
		expectedHash := a.calculateHash(entry)
		if entry.Hash != expectedHash {
			return fmt.Errorf("hash mismatch at entry %s: expected %s, got %s", entry.ID, expectedHash, entry.Hash)
		}

		prevHash = entry.Hash
	}

	return nil
}

// Cleanup removes old audit logs beyond retention period
func (a *AuditLogger) Cleanup() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -a.retentionDays)

	entries, err := os.ReadDir(a.auditDir)
	if err != nil {
		return fmt.Errorf("failed to read audit directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Parse date from filename
		dateStr := strings.TrimSuffix(entry.Name(), ".log")
		logDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue // Skip files that don't match pattern
		}

		if logDate.Before(cutoff) {
			path := filepath.Join(a.auditDir, entry.Name())
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to remove old log file %s: %w", path, err)
			}
		}
	}

	return nil
}

// Close flushes remaining entries and closes the logger
func (a *AuditLogger) Close() error {
	close(a.done)
	a.flushTicker.Stop()

	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.flushBuffer(); err != nil {
		return err
	}

	if a.currentFile != nil {
		return a.currentFile.Close()
	}

	return nil
}

// backgroundFlush periodically flushes the buffer
func (a *AuditLogger) backgroundFlush() {
	for {
		select {
		case <-a.flushTicker.C:
			a.mu.Lock()
			a.flushBuffer()
			a.mu.Unlock()
		case <-a.done:
			return
		}
	}
}

func (a *AuditLogger) flushBuffer() error {
	if len(a.buffer) == 0 {
		return nil
	}

	if time.Now().Format("2006-01-02") != a.currentDate {
		if err := a.rotateLogFile(); err != nil {
			return err
		}
	}

	for _, entry := range a.buffer {
		line, err := json.Marshal(entry)
		if err != nil {
			return fmt.Errorf("failed to marshal entry: %w", err)
		}

		if _, err := a.currentFile.Write(append(line, '\n')); err != nil {
			return fmt.Errorf("failed to write entry: %w", err)
		}
	}

	if err := a.currentFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	a.buffer = a.buffer[:0]
	a.bufferSize = 0

	return nil
}

func (a *AuditLogger) rotateLogFile() error {
	if a.currentFile != nil {
		a.currentFile.Close()
	}

	// Generate filename for today
	a.currentDate = time.Now().Format("2006-01-02")
	filename := filepath.Join(a.auditDir, a.currentDate+".log")

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	a.currentFile = file

	if err := a.loadLastHash(filename); err != nil {
		return err
	}

	return nil
}

func (a *AuditLogger) loadLastHash(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			a.lastHash = ""
			return nil
		}
		return err
	}

	lines := strings.Split(string(data), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		var entry AuditEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		a.lastHash = entry.Hash
		return nil
	}

	a.lastHash = ""
	return nil
}

func (a *AuditLogger) calculateHash(entry AuditEntry) string {
	h := sha256.New()
	h.Write([]byte(entry.ID))
	h.Write([]byte(entry.Timestamp.String()))
	h.Write([]byte(entry.Action))
	h.Write([]byte(entry.Target))
	h.Write([]byte(entry.Actor))
	h.Write([]byte(fmt.Sprintf("%v", entry.Success)))
	h.Write([]byte(entry.Error))
	h.Write([]byte(entry.PrevHash))

	if entry.Metadata != nil {
		metadataJSON, _ := json.Marshal(entry.Metadata)
		h.Write(metadataJSON)
	}

	return hex.EncodeToString(h.Sum(nil))
}

func (a *AuditLogger) getLogFiles(startTime, endTime *time.Time) ([]string, error) {
	entries, err := os.ReadDir(a.auditDir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		dateStr := strings.TrimSuffix(entry.Name(), ".log")
		logDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}

		if startTime != nil && logDate.Before(startTime.Truncate(24*time.Hour)) {
			continue
		}
		if endTime != nil && logDate.After(endTime.Truncate(24*time.Hour)) {
			continue
		}

		files = append(files, filepath.Join(a.auditDir, entry.Name()))
	}

	return files, nil
}

func (a *AuditLogger) readLogFile(filename string, opts QueryOptions) ([]AuditEntry, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var results []AuditEntry
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var entry AuditEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		if !a.matchesFilters(entry, opts) {
			continue
		}

		results = append(results, entry)
	}

	return results, nil
}

func (a *AuditLogger) matchesFilters(entry AuditEntry, opts QueryOptions) bool {
	if opts.StartTime != nil && entry.Timestamp.Before(*opts.StartTime) {
		return false
	}
	if opts.EndTime != nil && entry.Timestamp.After(*opts.EndTime) {
		return false
	}

	if len(opts.Actions) > 0 {
		found := false
		for _, action := range opts.Actions {
			if entry.Action == action {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if opts.Actor != "" && entry.Actor != opts.Actor {
		return false
	}

	if opts.Target != "" && entry.Target != opts.Target {
		return false
	}

	if opts.Success != nil && entry.Success != *opts.Success {
		return false
	}

	return true
}

func (a *AuditLogger) exportCSV(entries []AuditEntry, w *os.File) error {
	header := "timestamp,action,target,actor,success,error\n"
	if _, err := w.WriteString(header); err != nil {
		return err
	}

	for _, entry := range entries {
		line := fmt.Sprintf("%s,%s,%s,%s,%v,%q\n",
			entry.Timestamp.Format(time.RFC3339),
			entry.Action,
			entry.Target,
			entry.Actor,
			entry.Success,
			entry.Error,
		)
		if _, err := w.WriteString(line); err != nil {
			return err
		}
	}

	return nil
}
