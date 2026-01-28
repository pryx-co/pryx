package mcp

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"sync"

	"github.com/playwright-community/playwright-go"
)

type BrowserProvider struct {
	mu      sync.Mutex
	started bool

	pw      *playwright.Playwright
	browser playwright.Browser
	page    playwright.Page
}

func NewBrowserProvider() *BrowserProvider {
	return &BrowserProvider{}
}

func (p *BrowserProvider) ServerInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":    "pryx-core/browser",
		"title":   "Pryx Browser (Bundled)",
		"version": "dev",
	}
}

func (p *BrowserProvider) ListTools(ctx context.Context) ([]Tool, error) {
	_ = ctx
	return []Tool{
		{Name: "install", Title: "Install Playwright", InputSchema: schemaRaw(`{"type":"object","properties":{"browsers":{"type":"array","items":{"type":"string"},"default":["chromium"]}},"additionalProperties":false}`)},
		{Name: "goto", Title: "Navigate", InputSchema: schemaRaw(`{"type":"object","properties":{"url":{"type":"string"}},"required":["url"],"additionalProperties":false}`)},
		{Name: "content", Title: "Get Page HTML", InputSchema: schemaRaw(`{"type":"object","properties":{},"additionalProperties":false}`)},
		{Name: "screenshot", Title: "Screenshot", InputSchema: schemaRaw(`{"type":"object","properties":{"full_page":{"type":"boolean","default":false}},"additionalProperties":false}`)},
		{Name: "evaluate", Title: "Evaluate JavaScript", InputSchema: schemaRaw(`{"type":"object","properties":{"expression":{"type":"string"}},"required":["expression"],"additionalProperties":false}`)},
	}, nil
}

func (p *BrowserProvider) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (ToolResult, error) {
	switch name {
	case "install":
		return p.install(ctx, arguments)
	case "goto":
		return p.gotoURL(ctx, arguments)
	case "content":
		return p.content(ctx)
	case "screenshot":
		return p.screenshot(ctx, arguments)
	case "evaluate":
		return p.evaluate(ctx, arguments)
	default:
		return ToolResult{}, errors.New("unknown tool")
	}
}

func (p *BrowserProvider) install(ctx context.Context, arguments map[string]interface{}) (ToolResult, error) {
	_ = ctx
	browsers := []string{"chromium"}
	if raw, ok := arguments["browsers"]; ok {
		if arr, ok := raw.([]interface{}); ok && len(arr) > 0 {
			var out []string
			for _, v := range arr {
				if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
					out = append(out, strings.TrimSpace(s))
				}
			}
			if len(out) > 0 {
				browsers = out
			}
		}
	}
	if err := playwright.Install(&playwright.RunOptions{Browsers: browsers}); err != nil {
		return ToolResult{}, err
	}
	return ToolResult{Content: []ToolContent{{Type: "text", Text: "OK"}}}, nil
}

func (p *BrowserProvider) ensureStarted(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.started {
		return nil
	}

	pw, err := playwright.Run()
	if err != nil {
		return err
	}
	b, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		_ = pw.Stop()
		return err
	}
	page, err := b.NewPage()
	if err != nil {
		_ = b.Close()
		_ = pw.Stop()
		return err
	}

	_ = ctx
	p.pw = pw
	p.browser = b
	p.page = page
	p.started = true
	return nil
}

func (p *BrowserProvider) gotoURL(ctx context.Context, arguments map[string]interface{}) (ToolResult, error) {
	url := strings.TrimSpace(argString(arguments, "url"))
	if url == "" {
		return ToolResult{}, errors.New("missing url")
	}
	if err := p.ensureStarted(ctx); err != nil {
		return ToolResult{}, err
	}
	_, err := p.page.Goto(url)
	if err != nil {
		return ToolResult{}, err
	}
	return ToolResult{Content: []ToolContent{{Type: "text", Text: "OK"}}}, nil
}

func (p *BrowserProvider) content(ctx context.Context) (ToolResult, error) {
	if err := p.ensureStarted(ctx); err != nil {
		return ToolResult{}, err
	}
	html, err := p.page.Content()
	if err != nil {
		return ToolResult{}, err
	}
	return ToolResult{
		Content:           []ToolContent{{Type: "text", Text: "OK"}},
		StructuredContent: jsonRaw(map[string]interface{}{"html": html}),
	}, nil
}

func (p *BrowserProvider) screenshot(ctx context.Context, arguments map[string]interface{}) (ToolResult, error) {
	if err := p.ensureStarted(ctx); err != nil {
		return ToolResult{}, err
	}
	fullPage := argBool(arguments, "full_page")
	data, err := p.page.Screenshot(playwright.PageScreenshotOptions{FullPage: playwright.Bool(fullPage)})
	if err != nil {
		return ToolResult{}, err
	}
	enc := base64.StdEncoding.EncodeToString(data)
	return ToolResult{
		Content: []ToolContent{
			{Type: "image", Data: enc, MimeType: "image/png"},
		},
	}, nil
}

func (p *BrowserProvider) evaluate(ctx context.Context, arguments map[string]interface{}) (ToolResult, error) {
	expr := strings.TrimSpace(argString(arguments, "expression"))
	if expr == "" {
		return ToolResult{}, errors.New("missing expression")
	}
	if err := p.ensureStarted(ctx); err != nil {
		return ToolResult{}, err
	}
	res, err := p.page.Evaluate(expr)
	if err != nil {
		return ToolResult{}, err
	}
	return ToolResult{
		Content:           []ToolContent{{Type: "text", Text: "OK"}},
		StructuredContent: jsonRaw(map[string]interface{}{"result": res}),
	}, nil
}
