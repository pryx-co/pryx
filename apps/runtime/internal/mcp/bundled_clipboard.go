package mcp

import (
	"context"
	"encoding/base64"
	"errors"
	"os/exec"
	"runtime"
	"strings"
)

type ClipboardProvider struct{}

func NewClipboardProvider() *ClipboardProvider {
	return &ClipboardProvider{}
}

func (p *ClipboardProvider) ServerInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":    "pryx-core/clipboard",
		"title":   "Pryx Clipboard (Bundled)",
		"version": "dev",
	}
}

func (p *ClipboardProvider) ListTools(ctx context.Context) ([]Tool, error) {
	_ = ctx
	return []Tool{
		{Name: "read_clipboard", Title: "Read Clipboard", InputSchema: schemaRaw(`{"type":"object","properties":{"format":{"type":"string","enum":["text","base64"],"default":"text"}},"additionalProperties":false}`)},
		{Name: "write_clipboard", Title: "Write Clipboard", InputSchema: schemaRaw(`{"type":"object","properties":{"content":{"type":"string"},"format":{"type":"string","enum":["text","base64"],"default":"text"}},"required":["content"],"additionalProperties":false}`)},
	}, nil
}

func (p *ClipboardProvider) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (ToolResult, error) {
	_ = ctx
	switch name {
	case "read_clipboard":
		return p.readClipboard(arguments)
	case "write_clipboard":
		return p.writeClipboard(arguments)
	default:
		return ToolResult{}, errors.New("unknown tool")
	}
}

func (p *ClipboardProvider) readClipboard(arguments map[string]interface{}) (ToolResult, error) {
	format := strings.ToLower(strings.TrimSpace(argString(arguments, "format")))
	if format == "" {
		format = "text"
	}

	content, err := getClipboardContent()
	if err != nil {
		return ToolResult{}, err
	}

	out := map[string]interface{}{
		"content": content,
	}

	if format == "base64" {
		encoded := base64.StdEncoding.EncodeToString([]byte(content))
		out["content"] = encoded
		out["format"] = "base64"
	} else {
		out["format"] = "text"
	}

	return ToolResult{
		Content:           []ToolContent{{Type: "text", Text: "OK"}},
		StructuredContent: jsonRaw(out),
	}, nil
}

func (p *ClipboardProvider) writeClipboard(arguments map[string]interface{}) (ToolResult, error) {
	content := strings.TrimSpace(argString(arguments, "content"))
	if content == "" {
		return ToolResult{}, errors.New("missing content")
	}

	format := strings.ToLower(strings.TrimSpace(argString(arguments, "format")))
	if format == "" {
		format = "text"
	}

	var data string
	if format == "base64" {
		decoded, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			return ToolResult{}, err
		}
		data = string(decoded)
	} else {
		data = content
	}

	if err := setClipboardContent(data); err != nil {
		return ToolResult{}, err
	}

	return ToolResult{
		Content:           []ToolContent{{Type: "text", Text: "OK"}},
		StructuredContent: jsonRaw(map[string]interface{}{"format": format}),
	}, nil
}

func getClipboardContent() (string, error) {
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("pbpaste")
		output, err := cmd.Output()
		if err != nil {
			return "", err
		}
		return string(output), nil
	} else if runtime.GOOS == "linux" {
		cmd := exec.Command("xclip", "-selection", "clipboard", "-o")
		output, err := cmd.Output()
		if err != nil {
			return "", err
		}
		return string(output), nil
	} else if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", "-Command", "Get-Clipboard")
		output, err := cmd.Output()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(output)), nil
	}
	return "", errors.New("unsupported platform")
}

func setClipboardContent(content string) error {
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(content)
		return cmd.Run()
	} else if runtime.GOOS == "linux" {
		cmd := exec.Command("xclip", "-selection", "clipboard")
		cmd.Stdin = strings.NewReader(content)
		return cmd.Run()
	} else if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", "-Command", "Set-Clipboard -Value "+content)
		return cmd.Run()
	}
	return errors.New("unsupported platform")
}
