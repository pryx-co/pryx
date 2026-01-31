package mcp

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type ScreenProvider struct {
	captureDir string
}

func NewScreenProvider() *ScreenProvider {
	home, _ := os.UserHomeDir()
	captureDir := filepath.Join(home, ".pryx", "captures")
	os.MkdirAll(captureDir, 0755)
	return &ScreenProvider{captureDir: captureDir}
}

func (p *ScreenProvider) ServerInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":    "pryx-core/screen",
		"title":   "Pryx Screen Capture (Bundled)",
		"version": "dev",
	}
}

func (p *ScreenProvider) ListTools(ctx context.Context) ([]Tool, error) {
	_ = ctx
	tools := []Tool{
		{
			Name:        "capture",
			Title:       "Capture Screenshot",
			Description: "Capture a screenshot of the entire screen or a specific region",
			InputSchema: schemaRaw(`{"type":"object","properties":{"region":{"type":"string","description":"Region to capture (full, active, or x,y,width,height)"},"format":{"type":"string","enum":["png","jpg"],"default":"png"}},"required":["region"]}`),
		},
		{
			Name:        "record",
			Title:       "Record Screen",
			Description: "Record the screen for a specified duration (requires ffmpeg)",
			InputSchema: schemaRaw(`{"type":"object","properties":{"duration":{"type":"integer","description":"Recording duration in seconds"},"region":{"type":"string","description":"Region to record (full, active, or x,y,width,height)"},"fps":{"type":"integer","default":30},"format":{"type":"string","enum":["mp4","gif"],"default":"mp4"}},"required":["duration"]}`),
		},
	}
	return tools, nil
}

func (p *ScreenProvider) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (ToolResult, error) {
	switch name {
	case "capture":
		return p.captureScreen(ctx, arguments)
	case "record":
		return p.recordScreen(ctx, arguments)
	default:
		return ToolResult{}, fmt.Errorf("unknown tool: %s", name)
	}
}

func (p *ScreenProvider) captureScreen(ctx context.Context, args map[string]interface{}) (ToolResult, error) {
	region, _ := args["region"].(string)
	format, _ := args["format"].(string)
	if format == "" {
		format = "png"
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("capture_%s.%s", timestamp, format)
	filepath := filepath.Join(p.captureDir, filename)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = p.captureMacOS(region, filepath, format)
	case "linux":
		cmd = p.captureLinux(region, filepath, format)
	case "windows":
		cmd = p.captureWindows(region, filepath, format)
	default:
		return ToolResult{}, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if cmd == nil {
		return ToolResult{}, fmt.Errorf("failed to create capture command")
	}

	cmdCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd = exec.CommandContext(cmdCtx, cmd.Path, cmd.Args[1:]...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return ToolResult{
			Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Capture failed: %v\nOutput: %s", err, string(output))}},
			IsError: true,
		}, nil
	}

	content := []ToolContent{
		{Type: "text", Text: fmt.Sprintf("Screenshot saved to: %s", filepath)},
	}

	if data, err := os.ReadFile(filepath); err == nil {
		mimeType := "image/png"
		if format == "jpg" {
			mimeType = "image/jpeg"
		}
		content = append(content, ToolContent{
			Type:     "image",
			Text:     filename,
			Data:     base64.StdEncoding.EncodeToString(data),
			MimeType: mimeType,
		})
	}

	return ToolResult{Content: content}, nil
}

func (p *ScreenProvider) captureMacOS(region, filepath, format string) *exec.Cmd {
	if region == "full" || region == "" {
		return exec.Command("screencapture", "-x", filepath)
	}
	return exec.Command("screencapture", "-x", "-R", region, filepath)
}

func (p *ScreenProvider) captureLinux(region, filepath, format string) *exec.Cmd {
	if _, err := exec.LookPath("gnome-screenshot"); err == nil {
		if region == "full" || region == "" {
			return exec.Command("gnome-screenshot", "-f", filepath)
		}
		return exec.Command("gnome-screenshot", "-a", "-f", filepath)
	}
	if _, err := exec.LookPath("import"); err == nil {
		if region == "full" || region == "" {
			return exec.Command("import", "-window", "root", filepath)
		}
		return exec.Command("import", filepath)
	}
	return nil
}

func (p *ScreenProvider) captureWindows(region, filepath, format string) *exec.Cmd {
	return nil
}

func (p *ScreenProvider) recordScreen(ctx context.Context, args map[string]interface{}) (ToolResult, error) {
	durationFloat, ok := args["duration"].(float64)
	if !ok {
		return ToolResult{}, fmt.Errorf("duration must be a number")
	}
	duration := int(durationFloat)

	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return ToolResult{
			Content: []ToolContent{{
				Type: "text",
				Text: "Screen recording requires ffmpeg. Install it:\n- macOS: brew install ffmpeg\n- Ubuntu/Debian: sudo apt install ffmpeg\n- Windows: choco install ffmpeg",
			}},
			IsError: true,
		}, nil
	}

	region, _ := args["region"].(string)
	fpsFloat, _ := args["fps"].(float64)
	fps := int(fpsFloat)
	if fps == 0 {
		fps = 30
	}
	format, _ := args["format"].(string)
	if format == "" {
		format = "mp4"
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("recording_%s.%s", timestamp, format)
	filepath := filepath.Join(p.captureDir, filename)

	display := "0"
	if runtime.GOOS == "darwin" {
		display = "1"
	}

	var cmd *exec.Cmd
	if region == "full" || region == "" {
		cmd = exec.Command("ffmpeg",
			"-f", "avfoundation",
			"-i", display,
			"-t", fmt.Sprintf("%d", duration),
			"-r", fmt.Sprintf("%d", fps),
			"-pix_fmt", "yuv420p",
			filepath,
		)
	} else {
		parts := strings.Split(region, ",")
		if len(parts) == 4 {
			x, y, w, h := parts[0], parts[1], parts[2], parts[3]
			cmd = exec.Command("ffmpeg",
				"-f", "avfoundation",
				"-i", display,
				"-vf", fmt.Sprintf("crop=%s:%s:%s:%s", w, h, x, y),
				"-t", fmt.Sprintf("%d", duration),
				"-r", fmt.Sprintf("%d", fps),
				"-pix_fmt", "yuv420p",
				filepath,
			)
		}
	}

	if cmd == nil {
		return ToolResult{}, fmt.Errorf("failed to create recording command")
	}

	cmdCtx, cancel := context.WithTimeout(ctx, time.Duration(duration+10)*time.Second)
	defer cancel()
	cmd = exec.CommandContext(cmdCtx, cmd.Path, cmd.Args[1:]...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return ToolResult{
			Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Recording failed: %v\nOutput: %s", err, string(output))}},
			IsError: true,
		}, nil
	}

	return ToolResult{
		Content: []ToolContent{{
			Type: "text",
			Text: fmt.Sprintf("Screen recording saved to: %s\nDuration: %d seconds\nFPS: %d", filepath, duration, fps),
		}},
	}, nil
}
