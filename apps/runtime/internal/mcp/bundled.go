package mcp

import (
	"errors"
	"strings"
)

func BundledProvider(name string) (ToolProvider, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "filesystem", "fs":
		return NewFilesystemProvider(), nil
	case "shell", "sh":
		return NewShellProvider(), nil
	case "clipboard":
		return NewClipboardProvider(), nil
	case "browser":
		return NewBrowserProvider(), nil
	default:
		return nil, errors.New("unknown bundled server")
	}
}
