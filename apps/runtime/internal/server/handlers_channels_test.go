package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleChannelTypes(t *testing.T) {
	handler := createTestHandler()

	req := httptest.NewRequest("GET", "/api/v1/channels/types", nil)
	w := httptest.NewRecorder()

	handler.handleChannelTypes(w, req)

	rec := w.Result()

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if types, ok := result["types"].([]interface{}); ok {
		if len(types) != 4 {
			t.Errorf("expected 4 channel types, got %d", len(types))
		}
	} else {
		t.Error("expected types in response")
	}
}

func TestHandleChannelCreate(t *testing.T) {
	handler := createTestHandler()

	tests := []struct {
		name    string
		channel string
		wantCode int
	}{
		{
			name:    "telegram channel",
			channel: `{"type":"telegram","name":"Test","config":{"token_ref":"test"}}`,
			wantCode: http.StatusOK,
		},
		{
			name:    "missing type",
			channel: `{"name":"Test"}`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:    "missing name",
			channel: `{"type":"slack"}`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:    "invalid type",
			channel: `{"type":"invalid","name":"Test"}`,
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/channels", nil)
			w := httptest.NewRecorder()

			handler.handleChannelCreate(w, req)

			rec := w.Result()

			if rec.Code != tt.wantCode {
				t.Errorf("expected status %d, got %d", tt.wantCode, rec.Code)
			}
		})
	}
}

func TestHandleChannelUpdate(t *testing.T) {
	handler := createTestHandler()

	body := json.Marshal(map[string]string{"type": "telegram", "name": "Updated Name"})

	req := httptest.NewRequest("PUT", "/api/v1/channels/test-id", body)
	w := httptest.NewRecorder()

	handler.handleChannelUpdate(w, req)

	rec := w.Result()

	if rec.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.StatusCode)
	}
}

func TestHandleChannelDelete(t *testing.T) {
	handler := createTestHandler()

	req := httptest.NewRequest("DELETE", "/api/v1/channels/test-id", nil)
	w := httptest.NewRecorder()

	handler.handleChannelDelete(w, req)

	rec := w.Result()

	if rec.StatusCode != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.StatusCode)
	}
}

func TestHandleChannelHealth(t *testing.T) {
	handler := createTestHandler()

	req := httptest.NewRequest("GET", "/api/v1/channels/test-id/health", nil)
	w := httptest.NewRecorder()

	handler.handleChannelHealth(w, req)

	rec := w.Result()

	if rec.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.StatusCode)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if healthy, ok := result["healthy"].(bool); ok {
		if !healthy {
			t.Error("expected healthy to be true")
		}
	}
}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if types, ok := result["types"].([]interface{}); ok {
		if len(types) != 4 {
			t.Errorf("expected 4 channel types, got %d", len(types))
		}
	} else {
		t.Error("expected types in response")
	}
}

func TestHandleChannelCreate(t *testing.T) {
	handler := createTestHandler()

	tests := []struct {
		name     string
		channel  string
		wantCode int
	}{
		{
			name:     "telegram channel",
			channel:  `{"type":"telegram","name":"Test","config":{"token_ref":"test"}}`,
			wantCode: http.StatusOK,
		},
		{
			name:     "missing type",
			channel:  `{"name":"Test"}`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "missing name",
			channel:  `{"type":"slack"}`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "invalid type",
			channel:  `{"type":"invalid","name":"Test"}`,
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/channels", nil)
			w := httptest.NewRecorder()

			handler.handleChannelCreate(w, req)

			rec := w.Result()

			if rec.Code != tt.wantCode {
				t.Errorf("expected status %d, got %d", tt.wantCode, rec.Code)
			}
		})
	}
}

func TestHandleChannelUpdate(t *testing.T) {
	handler := createTestHandler()

	body := json.Marshal(map[string]string{"type": "telegram", "name": "Updated Name"})

	req := httptest.NewRequest("PUT", "/api/v1/channels/test-id", body)
	w := httptest.NewRecorder()

	handler.handleChannelUpdate(w, req)

	rec := w.Result()

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestHandleChannelDelete(t *testing.T) {
	handler := createTestHandler()

	req := httptest.NewRequest("DELETE", "/api/v1/channels/test-id", nil)
	w := httptest.NewRecorder()

	handler.handleChannelDelete(w, req)

	rec := w.Result()

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
}

func TestHandleChannelHealth(t *testing.T) {
	handler := createTestHandler()

	req := httptest.NewRequest("GET", "/api/v1/channels/test-id/health", nil)
	w := httptest.NewRecorder()

	handler.handleChannelHealth(w, req)

	rec := w.Result()

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if healthy, ok := result["healthy"].(bool); ok {
		if !healthy {
			t.Error("expected healthy to be true")
		}
	}
}

func createTestHandler() *Server {
	s := &Server{}
	s.channels = nil
	s.routes()
	return s
}
