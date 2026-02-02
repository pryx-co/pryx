package server

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleSessionsList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if s.store == nil {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"sessions": []interface{}{}})
		return
	}
	sessions, err := s.store.ListSessions()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	resp := make([]map[string]interface{}, 0, len(sessions))
	for _, sess := range sessions {
		resp = append(resp, map[string]interface{}{
			"id":        sess.ID,
			"title":     sess.Title,
			"createdAt": sess.CreatedAt.Format(timeRFC3339),
			"updatedAt": sess.UpdatedAt.Format(timeRFC3339),
		})
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"sessions": resp})
}

func (s *Server) handleSessionCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Name  string `json:"name"`
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	if s.store == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "store not available"})
		return
	}

	title := req.Title
	if title == "" {
		title = req.Name
	}
	if title == "" {
		title = "Session"
	}

	sess, err := s.store.CreateSession(title)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"id":        sess.ID,
		"title":     sess.Title,
		"createdAt": sess.CreatedAt.Format(timeRFC3339),
		"updatedAt": sess.UpdatedAt.Format(timeRFC3339),
	})
}

func (s *Server) handleSessionGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "session id is required"})
		return
	}

	if s.store == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "store not available"})
		return
	}

	sess, err := s.store.GetSession(sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	msgCount, _ := s.store.GetMessageCount(sessionID)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"id":           sess.ID,
		"title":        sess.Title,
		"createdAt":    sess.CreatedAt.Format(timeRFC3339),
		"updatedAt":    sess.UpdatedAt.Format(timeRFC3339),
		"messageCount": msgCount,
	})
}

func (s *Server) handleSessionDelete(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "session id is required"})
		return
	}

	if s.store == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "store not available"})
		return
	}

	if err := s.store.DeleteSession(sessionID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

const timeRFC3339 = "2006-01-02T15:04:05Z07:00"
