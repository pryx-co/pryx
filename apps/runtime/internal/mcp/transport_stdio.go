package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

type StdioTransport struct {
	command []string
	cwd     string
	env     map[string]string

	startOnce sync.Once
	startErr  error

	procCancel context.CancelFunc

	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser

	mu      sync.Mutex
	pending map[string]chan RPCResponse
	closed  bool
}

func NewStdioTransport(command []string, cwd string, env map[string]string) *StdioTransport {
	return &StdioTransport{
		command: command,
		cwd:     cwd,
		env:     env,
		pending: map[string]chan RPCResponse{},
	}
}

func (t *StdioTransport) Start(ctx context.Context) error {
	t.startOnce.Do(func() {
		if len(t.command) == 0 {
			t.startErr = errors.New("missing command")
			return
		}

		procCtx, cancel := context.WithCancel(context.Background())
		t.procCancel = cancel

		t.cmd = exec.CommandContext(procCtx, t.command[0], t.command[1:]...)
		if t.cwd != "" {
			t.cmd.Dir = t.cwd
		}
		if len(t.env) > 0 {
			var out []string
			for k, v := range t.env {
				out = append(out, fmt.Sprintf("%s=%s", k, v))
			}
			t.cmd.Env = append(t.cmd.Environ(), out...)
		}

		stdout, err := t.cmd.StdoutPipe()
		if err != nil {
			t.startErr = err
			return
		}
		stdin, err := t.cmd.StdinPipe()
		if err != nil {
			t.startErr = err
			return
		}

		if err := t.cmd.Start(); err != nil {
			t.startErr = err
			return
		}

		t.stdin = stdin
		t.stdout = stdout

		go t.readLoop()
		go func() {
			_ = t.cmd.Wait()
			t.failAll(errors.New("stdio server exited"))
		}()
	})
	return t.startErr
}

func (t *StdioTransport) Close() error {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return nil
	}
	t.closed = true

	for _, ch := range t.pending {
		close(ch)
	}
	t.pending = map[string]chan RPCResponse{}
	t.mu.Unlock()

	if t.stdin != nil {
		_ = t.stdin.Close()
	}
	if t.stdout != nil {
		_ = t.stdout.Close()
	}
	if t.procCancel != nil {
		t.procCancel()
	}
	if t.cmd != nil && t.cmd.Process != nil {
		_ = t.cmd.Process.Kill()
	}
	return nil
}

func (t *StdioTransport) Call(ctx context.Context, req RPCRequest) (RPCResponse, error) {
	if err := t.Start(ctx); err != nil {
		return RPCResponse{}, err
	}

	b, err := json.Marshal(req)
	if err != nil {
		return RPCResponse{}, err
	}

	var idRaw json.RawMessage
	if req.ID != nil {
		idRaw, _ = json.Marshal(req.ID)
	}
	key := idKey(idRaw)
	if key == "" {
		return RPCResponse{}, errors.New("missing id")
	}

	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return RPCResponse{}, errors.New("transport closed")
	}
	ch := make(chan RPCResponse, 1)
	t.pending[key] = ch
	t.mu.Unlock()

	if _, err := t.stdin.Write(append(b, '\n')); err != nil {
		t.mu.Lock()
		delete(t.pending, key)
		t.mu.Unlock()
		return RPCResponse{}, err
	}

	select {
	case <-ctx.Done():
		t.mu.Lock()
		delete(t.pending, key)
		t.mu.Unlock()
		return RPCResponse{}, ctx.Err()
	case resp, ok := <-ch:
		if !ok {
			return RPCResponse{}, errors.New("transport closed")
		}
		return resp, nil
	}
}

func (t *StdioTransport) Notify(ctx context.Context, notif RPCNotification) error {
	if err := t.Start(ctx); err != nil {
		return err
	}
	b, err := json.Marshal(notif)
	if err != nil {
		return err
	}
	_, err = t.stdin.Write(append(b, '\n'))
	return err
}

func (t *StdioTransport) readLoop() {
	scanner := bufio.NewScanner(t.stdout)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		resp := RPCResponse{}
		if err := json.Unmarshal(line, &resp); err != nil {
			continue
		}
		key := idKey(resp.ID)
		if key == "" {
			continue
		}

		t.mu.Lock()
		ch, ok := t.pending[key]
		if ok {
			delete(t.pending, key)
		}
		t.mu.Unlock()

		if ok {
			ch <- resp
			close(ch)
		}
	}

	if err := scanner.Err(); err != nil {
		t.failAll(err)
	} else {
		t.failAll(errors.New("stdio stream ended"))
	}
}

func (t *StdioTransport) failAll(err error) {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return
	}
	t.closed = true
	for _, ch := range t.pending {
		close(ch)
	}
	t.pending = map[string]chan RPCResponse{}
	t.mu.Unlock()

	if t.procCancel != nil {
		t.procCancel()
	}
	_ = err
}
