package agentbus

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// StructuredLogger provides structured JSON logging with correlation IDs
type StructuredLogger struct {
	name     string
	level    string
	mu       sync.Mutex
	instance *log.Logger
	output   *os.File
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger(name, level string) *StructuredLogger {
	logger := &StructuredLogger{
		name:  name,
		level: level,
	}

	// Ensure logs directory exists
	logsDir := filepath.Join(os.Getenv("HOME"), ".pryx", "logs")
	if err := os.MkdirAll(logsDir, 0755); err == nil {
		logFile, err := os.OpenFile(
			filepath.Join(logsDir, "agentbus.log"),
			os.O_CREATE|os.O_WRONLY|os.O_APPEND,
			0644,
		)
		if err == nil {
			logger.output = logFile
			logger.instance = log.New(logFile, "", 0)
		}
	}

	if logger.instance == nil {
		logger.instance = log.New(os.Stdout, "", 0)
	}

	return logger
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp  time.Time              `json:"timestamp"`
	Level      string                 `json:"level"`
	Service    string                 `json:"service"`
	Message    string                 `json:"message"`
	TraceID    string                 `json:"trace_id,omitempty"`
	SpanID     string                 `json:"span_id,omitempty"`
	AgentID    string                 `json:"agent_id,omitempty"`
	Protocol   string                 `json:"protocol,omitempty"`
	Action     string                 `json:"action,omitempty"`
	Error      string                 `json:"error,omitempty"`
	DurationMs int64                  `json:"duration_ms,omitempty"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
}

// Log logs a message with structured fields
func (l *StructuredLogger) Log(level, message string, fields map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().UTC(),
		Level:     level,
		Service:   l.name,
		Message:   message,
		Fields:    fields,
	}

	// Extract correlation IDs
	if traceID, ok := fields["trace_id"].(string); ok {
		entry.TraceID = traceID
	}
	if agentID, ok := fields["agent_id"].(string); ok {
		entry.AgentID = agentID
	}
	if protocol, ok := fields["protocol"].(string); ok {
		entry.Protocol = protocol
	}
	if action, ok := fields["action"].(string); ok {
		entry.Action = action
	}
	if err, ok := fields["error"].(string); ok {
		entry.Error = err
	}

	jsonBytes, _ := json.Marshal(entry)
	l.instance.Println(string(jsonBytes))
}

// Debug logs a debug message
func (l *StructuredLogger) Debug(message string, fields map[string]interface{}) {
	if l.shouldLog("debug") {
		l.Log("debug", message, fields)
	}
}

// Info logs an info message
func (l *StructuredLogger) Info(message string, fields map[string]interface{}) {
	if l.shouldLog("info") {
		l.Log("info", message, fields)
	}
}

// Warn logs a warning message
func (l *StructuredLogger) Warn(message string, fields map[string]interface{}) {
	if l.shouldLog("warn") {
		l.Log("warn", message, fields)
	}
}

// Error logs an error message
func (l *StructuredLogger) Error(message string, fields map[string]interface{}) {
	if l.shouldLog("error") {
		l.Log("error", message, fields)
	}
}

// Fatal logs a fatal message and exits
func (l *StructuredLogger) Fatal(message string, fields map[string]interface{}) {
	l.Log("fatal", message, fields)
	os.Exit(1)
}

// Trace spans for distributed tracing
type TraceSpan struct {
	TraceID    string                 `json:"trace_id"`
	SpanID     string                 `json:"span_id"`
	ParentID   string                 `json:"parent_id"`
	Operation  string                 `json:"operation"`
	StartTime  time.Time              `json:"start_time"`
	EndTime    *time.Time             `json:"end_time,omitempty"`
	DurationMs int64                  `json:"duration_ms,omitempty"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
	Status     string                 `json:"status"`
	Error      string                 `json:"error,omitempty"`
}

// NewTraceSpan creates a new trace span
func NewTraceSpan(traceID, parentID, operation string) *TraceSpan {
	return &TraceSpan{
		TraceID:   traceID,
		SpanID:    uuid.New().String(),
		ParentID:  parentID,
		Operation: operation,
		StartTime: time.Now().UTC(),
		Fields:    make(map[string]interface{}),
		Status:    "started",
	}
}

// Finish marks the span as complete
func (s *TraceSpan) Finish(fields map[string]interface{}) {
	now := time.Now()
	s.EndTime = &now
	s.Status = "completed"
	s.Fields = fields
	s.DurationMs = now.Sub(s.StartTime).Milliseconds()
}

// RecordError records an error in the span
func (s *TraceSpan) RecordError(err error) {
	s.Status = "error"
	s.Error = err.Error()
}

// Trace provides distributed tracing context
type Trace struct {
	TraceID string
	Spans   []*TraceSpan
	mu      sync.RWMutex
}

// NewTrace creates a new trace
func NewTrace() *Trace {
	return &Trace{
		TraceID: uuid.New().String(),
		Spans:   make([]*TraceSpan, 0),
	}
}

// StartSpan starts a new span
func (t *Trace) StartSpan(operation string) *TraceSpan {
	t.mu.Lock()
	defer t.mu.Unlock()

	var parentID string
	if len(t.Spans) > 0 {
		parentID = t.Spans[len(t.Spans)-1].SpanID
	}

	span := NewTraceSpan(t.TraceID, parentID, operation)
	t.Spans = append(t.Spans, span)
	return span
}

// GetSpans returns all spans in the trace
func (t *Trace) GetSpans() []*TraceSpan {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Spans
}

// CorrelationID generates a new correlation ID
func CorrelationID() string {
	return uuid.New().String()
}

// shouldLog checks if a log level should be output
func (l *StructuredLogger) shouldLog(level string) bool {
	levels := map[string]int{
		"debug": 0,
		"info":  1,
		"warn":  2,
		"error": 3,
		"fatal": 4,
	}

	currentLevel, currentOk := levels[l.level]
	targetLevel, targetOk := levels[level]

	if !currentOk {
		return targetOk && targetLevel >= 1 // Default to info
	}

	return targetOk && targetLevel >= currentLevel
}

// Close closes the logger
func (l *StructuredLogger) Close() {
	if l.output != nil {
		l.output.Close()
	}
}

// PrettyPrint returns a human-readable representation
func (l *LogEntry) PrettyPrint() string {
	return fmt.Sprintf("LogEntry{Level: %s, Message: %s, TraceID: %s}",
		l.Level, l.Message, l.TraceID)
}
