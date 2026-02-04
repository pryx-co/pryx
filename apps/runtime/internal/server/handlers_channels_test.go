package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleChannelTypes(t *testing.T) {
	s := &Server{}
	s.channels = nil

	req := httptest.NewRequest("GET", "/api/v1/channels/types", nil)
	w := httptest.NewRecorder()

	s.handleChannelTypes(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
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
