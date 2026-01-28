package mcp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type FilesystemProvider struct {
	root string
}

func NewFilesystemProvider() *FilesystemProvider {
	root := strings.TrimSpace(os.Getenv("PRYX_WORKSPACE_ROOT"))
	if root == "" {
		if cwd, err := os.Getwd(); err == nil {
			root = cwd
		}
	}
	if root != "" {
		if abs, err := filepath.Abs(root); err == nil {
			root = abs
		}
	}
	return &FilesystemProvider{root: root}
}

func (p *FilesystemProvider) ServerInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":    "pryx-core/filesystem",
		"title":   "Pryx Filesystem (Bundled)",
		"version": "dev",
	}
}

func (p *FilesystemProvider) ListTools(ctx context.Context) ([]Tool, error) {
	_ = ctx
	return []Tool{
		{Name: "read_file", Title: "Read File", InputSchema: schemaRaw(`{"type":"object","properties":{"path":{"type":"string"},"encoding":{"type":"string","enum":["text","base64"],"default":"text"}},"required":["path"],"additionalProperties":false}`)},
		{Name: "write_file", Title: "Write File", InputSchema: schemaRaw(`{"type":"object","properties":{"path":{"type":"string"},"content":{"type":"string"},"encoding":{"type":"string","enum":["text","base64"],"default":"text"},"create_dirs":{"type":"boolean","default":false}},"required":["path","content"],"additionalProperties":false}`)},
		{Name: "list_dir", Title: "List Directory", InputSchema: schemaRaw(`{"type":"object","properties":{"path":{"type":"string"},"recursive":{"type":"boolean","default":false}},"required":["path"],"additionalProperties":false}`)},
		{Name: "mkdir", Title: "Make Directory", InputSchema: schemaRaw(`{"type":"object","properties":{"path":{"type":"string"},"parents":{"type":"boolean","default":true}},"required":["path"],"additionalProperties":false}`)},
		{Name: "remove", Title: "Remove File or Directory", InputSchema: schemaRaw(`{"type":"object","properties":{"path":{"type":"string"},"recursive":{"type":"boolean","default":false}},"required":["path"],"additionalProperties":false}`)},
	}, nil
}

func (p *FilesystemProvider) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (ToolResult, error) {
	_ = ctx
	switch name {
	case "read_file":
		path, err := p.argPath(arguments, "path")
		if err != nil {
			return ToolResult{}, err
		}
		encoding := strings.ToLower(strings.TrimSpace(argString(arguments, "encoding")))
		if encoding == "" {
			encoding = "text"
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return ToolResult{}, err
		}
		out := map[string]interface{}{
			"path": path,
		}
		if encoding == "base64" {
			out["content"] = base64.StdEncoding.EncodeToString(b)
			out["encoding"] = "base64"
		} else {
			out["content"] = string(b)
			out["encoding"] = "text"
		}
		return ToolResult{
			Content:           []ToolContent{{Type: "text", Text: "OK"}},
			StructuredContent: jsonRaw(out),
		}, nil

	case "write_file":
		path, err := p.argPath(arguments, "path")
		if err != nil {
			return ToolResult{}, err
		}
		content := argString(arguments, "content")
		encoding := strings.ToLower(strings.TrimSpace(argString(arguments, "encoding")))
		if encoding == "" {
			encoding = "text"
		}
		createDirs := argBool(arguments, "create_dirs")

		if createDirs {
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return ToolResult{}, err
			}
		}

		var b []byte
		if encoding == "base64" {
			decoded, err := base64.StdEncoding.DecodeString(content)
			if err != nil {
				return ToolResult{}, err
			}
			b = decoded
		} else {
			b = []byte(content)
		}

		if err := os.WriteFile(path, b, 0o600); err != nil {
			return ToolResult{}, err
		}
		return ToolResult{Content: []ToolContent{{Type: "text", Text: "OK"}}}, nil

	case "list_dir":
		path, err := p.argPath(arguments, "path")
		if err != nil {
			return ToolResult{}, err
		}
		recursive := argBool(arguments, "recursive")
		entries, err := p.listDir(path, recursive)
		if err != nil {
			return ToolResult{}, err
		}
		return ToolResult{
			Content:           []ToolContent{{Type: "text", Text: "OK"}},
			StructuredContent: jsonRaw(map[string]interface{}{"entries": entries}),
		}, nil

	case "mkdir":
		path, err := p.argPath(arguments, "path")
		if err != nil {
			return ToolResult{}, err
		}
		parents := true
		if _, ok := arguments["parents"]; ok {
			parents = argBool(arguments, "parents")
		}
		if parents {
			if err := os.MkdirAll(path, 0o755); err != nil {
				return ToolResult{}, err
			}
		} else {
			if err := os.Mkdir(path, 0o755); err != nil {
				return ToolResult{}, err
			}
		}
		return ToolResult{Content: []ToolContent{{Type: "text", Text: "OK"}}}, nil

	case "remove":
		path, err := p.argPath(arguments, "path")
		if err != nil {
			return ToolResult{}, err
		}
		recursive := argBool(arguments, "recursive")
		info, err := os.Stat(path)
		if err != nil {
			return ToolResult{}, err
		}
		if info.IsDir() {
			if !recursive {
				return ToolResult{}, errors.New("refusing to remove directory without recursive=true")
			}
			if err := os.RemoveAll(path); err != nil {
				return ToolResult{}, err
			}
		} else {
			if err := os.Remove(path); err != nil {
				return ToolResult{}, err
			}
		}
		return ToolResult{Content: []ToolContent{{Type: "text", Text: "OK"}}}, nil

	default:
		return ToolResult{}, errors.New("unknown tool")
	}
}

func (p *FilesystemProvider) argPath(args map[string]interface{}, key string) (string, error) {
	raw := strings.TrimSpace(argString(args, key))
	if raw == "" {
		return "", errors.New("missing path")
	}

	if strings.HasPrefix(raw, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			raw = filepath.Join(home, strings.TrimPrefix(raw, "~"))
		}
	}

	abs := raw
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(p.root, abs)
	}
	abs = filepath.Clean(abs)
	abs, err := filepath.Abs(abs)
	if err != nil {
		return "", err
	}

	root := p.root
	if root == "" {
		return abs, nil
	}
	root = filepath.Clean(root)
	if !strings.HasPrefix(abs, root+string(filepath.Separator)) && abs != root {
		return "", errors.New("path escapes workspace root")
	}
	return abs, nil
}

func (p *FilesystemProvider) listDir(path string, recursive bool) ([]map[string]interface{}, error) {
	var out []map[string]interface{}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		info, _ := e.Info()
		item := map[string]interface{}{
			"name":  e.Name(),
			"path":  filepath.Join(path, e.Name()),
			"isDir": e.IsDir(),
		}
		if info != nil {
			item["size"] = info.Size()
		}
		out = append(out, item)

		if recursive && e.IsDir() {
			child, err := p.listDir(filepath.Join(path, e.Name()), true)
			if err != nil {
				return nil, err
			}
			out = append(out, child...)
		}
	}
	return out, nil
}

func schemaRaw(s string) json.RawMessage {
	return json.RawMessage([]byte(s))
}

func jsonRaw(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func argString(args map[string]interface{}, key string) string {
	v, ok := args[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	b, _ := json.Marshal(v)
	return strings.Trim(string(b), `"`)
}

func argBool(args map[string]interface{}, key string) bool {
	v, ok := args[key]
	if !ok || v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	if s, ok := v.(string); ok {
		s = strings.ToLower(strings.TrimSpace(s))
		return s == "true" || s == "1" || s == "yes"
	}
	if f, ok := v.(float64); ok {
		return f != 0
	}
	return false
}
