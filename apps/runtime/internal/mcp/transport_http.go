package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

type HTTPTransport struct {
	url     string
	headers map[string]string
	client  *http.Client
}

func NewHTTPTransport(url string, headers map[string]string) *HTTPTransport {
	return &HTTPTransport{
		url:     url,
		headers: headers,
		client:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (t *HTTPTransport) Close() error {
	return nil
}

func (t *HTTPTransport) Call(ctx context.Context, req RPCRequest) (RPCResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return RPCResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, t.url, bytes.NewReader(body))
	if err != nil {
		return RPCResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	for k, v := range t.headers {
		if strings.TrimSpace(k) == "" {
			continue
		}
		httpReq.Header.Set(k, v)
	}

	resp, err := t.client.Do(httpReq)
	if err != nil {
		return RPCResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return RPCResponse{}, errors.New(strings.TrimSpace(string(b)))
	}

	ct := strings.ToLower(resp.Header.Get("Content-Type"))
	if strings.HasPrefix(ct, "text/event-stream") {
		return readSSEForResponse(resp.Body, req.ID)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return RPCResponse{}, err
	}
	out := RPCResponse{}
	if err := json.Unmarshal(b, &out); err != nil {
		return RPCResponse{}, err
	}
	return out, nil
}

func (t *HTTPTransport) Notify(ctx context.Context, notif RPCNotification) error {
	body, err := json.Marshal(notif)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, t.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	for k, v := range t.headers {
		if strings.TrimSpace(k) == "" {
			continue
		}
		httpReq.Header.Set(k, v)
	}

	resp, err := t.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted {
		return nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return errors.New(strings.TrimSpace(string(b)))
	}
	return nil
}

func readSSEForResponse(r io.Reader, reqID interface{}) (RPCResponse, error) {
	var idRaw json.RawMessage
	if reqID != nil {
		idRaw, _ = json.Marshal(reqID)
	}
	targetKey := idKey(idRaw)

	scanner := bufio.NewScanner(r)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 1024*1024)

	var dataLines []string
	flush := func() (*RPCResponse, bool) {
		if len(dataLines) == 0 {
			return nil, false
		}
		payload := strings.Join(dataLines, "\n")
		dataLines = nil
		payload = strings.TrimSpace(payload)
		if payload == "" {
			return nil, false
		}
		resp := RPCResponse{}
		if json.Unmarshal([]byte(payload), &resp) != nil {
			return nil, false
		}
		if targetKey == "" || idKey(resp.ID) == targetKey {
			return &resp, true
		}
		return nil, false
	}

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimRight(line, "\r")
		if line == "" {
			if resp, ok := flush(); ok {
				return *resp, nil
			}
			continue
		}
		if strings.HasPrefix(line, "data:") {
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}
	if resp, ok := flush(); ok {
		return *resp, nil
	}
	if err := scanner.Err(); err != nil {
		return RPCResponse{}, err
	}
	return RPCResponse{}, errors.New("no response in sse stream")
}
