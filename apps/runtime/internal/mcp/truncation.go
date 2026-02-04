package mcp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	MaxLines         = 2000
	MaxBytes         = 50 * 1024
	DefaultOutputDir = "tool-output"
	RetentionPeriod  = 7 * 24 * time.Hour
)

var (
	truncationMu    sync.RWMutex
	globalTruncator *Truncator
)

type TruncationResult struct {
	Content    string `json:"content"`
	Truncated  bool   `json:"truncated"`
	OutputPath string `json:"output_path,omitempty"`
}

type Truncator struct {
	outputDir string
	maxLines  int
	maxBytes  int
}

func NewTruncator(dataDir string) *Truncator {
	dir := filepath.Join(dataDir, DefaultOutputDir)
	return &Truncator{
		outputDir: dir,
		maxLines:  MaxLines,
		maxBytes:  MaxBytes,
	}
}

func (t *Truncator) Process(content string) (*TruncationResult, error) {
	lines := strings.Split(content, "\n")
	totalBytes := len(content)

	if len(lines) <= t.maxLines && totalBytes <= t.maxBytes {
		return &TruncationResult{
			Content:   content,
			Truncated: false,
		}, nil
	}

	if err := os.MkdirAll(t.outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	filename := fmt.Sprintf("tool_%d_%d.txt", time.Now().UnixNano(), os.Getpid())
	filepath := filepath.Join(t.outputDir, filename)

	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write output file: %w", err)
	}

	truncatedContent := t.truncateContent(lines, totalBytes)

	hint := fmt.Sprintf(
		"\n\n... output truncated (full content saved to: %s)\n"+
			"Use Grep to search the full content or Read with offset/limit to view specific sections.",
		filepath,
	)
	truncatedContent += hint

	return &TruncationResult{
		Content:    truncatedContent,
		Truncated:  true,
		OutputPath: filepath,
	}, nil
}

func (t *Truncator) truncateContent(lines []string, totalBytes int) string {
	var out []string
	var bytes int

	for i, line := range lines {
		if i >= t.maxLines {
			break
		}

		lineBytes := len(line)
		if i > 0 {
			lineBytes++
		}

		if bytes+lineBytes > t.maxBytes {
			break
		}

		out = append(out, line)
		bytes += lineBytes
	}

	return strings.Join(out, "\n")
}

func (t *Truncator) ProcessToolResult(result ToolResult) ToolResult {
	if result.IsError || len(result.Content) == 0 {
		return result
	}

	var newContent []ToolContent
	for _, content := range result.Content {
		if content.Type == "text" && content.Text != "" {
			truncated, err := t.Process(content.Text)
			if err != nil {
				newContent = append(newContent, content)
				continue
			}

			newContent = append(newContent, ToolContent{
				Type:     content.Type,
				Text:     truncated.Content,
				Data:     content.Data,
				MimeType: content.MimeType,
				URI:      content.URI,
				Name:     content.Name,
			})
		} else {
			newContent = append(newContent, content)
		}
	}

	return ToolResult{
		Content:           newContent,
		IsError:           result.IsError,
		StructuredContent: result.StructuredContent,
	}
}

func (t *Truncator) CleanupOldFiles() error {
	entries, err := os.ReadDir(t.outputDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	cutoff := time.Now().Add(-RetentionPeriod)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			path := filepath.Join(t.outputDir, entry.Name())
			os.Remove(path)
		}
	}

	return nil
}

func (t *Truncator) GetOutputDir() string {
	return t.outputDir
}

func (t *Truncator) SetOutputDir(dir string) {
	t.outputDir = dir
}

func InitTruncator(dataDir string) {
	truncationMu.Lock()
	defer truncationMu.Unlock()
	globalTruncator = NewTruncator(dataDir)
}

func GetTruncator() *Truncator {
	truncationMu.RLock()
	defer truncationMu.RUnlock()
	return globalTruncator
}

func TruncateToolResult(result ToolResult) ToolResult {
	truncator := GetTruncator()
	if truncator == nil {
		return result
	}
	return truncator.ProcessToolResult(result)
}

func Cleanup() error {
	truncator := GetTruncator()
	if truncator == nil {
		return nil
	}
	return truncator.CleanupOldFiles()
}
