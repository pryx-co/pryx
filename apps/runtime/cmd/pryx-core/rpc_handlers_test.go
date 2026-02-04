package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"pryx-core/internal/config"
	"pryx-core/internal/keychain"
	"pryx-core/internal/server"
	"pryx-core/internal/store"
)

func TestRPCHandlers(t *testing.T) {
	// Setup
	cfg := config.Load()
	db, _ := store.New(":memory:")
	kc := keychain.New("test")
	srv := server.New(cfg, db.DB, kc)
	registry := setupAdminHandlers(srv)

	tests := []struct {
		name   string
		method string
		params map[string]interface{}
		want   string
	}{
		{
			name:   "Health check",
			method: "admin.health",
			params: nil,
			want:   "healthy",
		},
		{
			name:   "Config get",
			method: "admin.config.get",
			params: nil,
			want:   "openai",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, ok := registry.GetHandler(tt.method)
			if !ok {
				t.Fatalf("Handler not found for %s", tt.method)
			}

			res, err := h(tt.method, tt.params)
			if err != nil {
				t.Fatalf("Handler failed: %v", err)
			}

			b, _ := json.Marshal(res)
			if !bytes.Contains(b, []byte(tt.want)) {
				t.Errorf("Response %s does not contain %s", string(b), tt.want)
			}
		})
	}
}
