package server

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"pryx-core/internal/bus"
	"pryx-core/internal/config"
	"pryx-core/internal/keychain"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"nhooyr.io/websocket"
)

type Server struct {
	cfg      *config.Config
	db       *sql.DB
	keychain *keychain.Keychain
	router   *chi.Mux
	bus      *bus.Bus
}

func New(cfg *config.Config, db *sql.DB, kc *keychain.Keychain) *Server {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	s := &Server{
		cfg:      cfg,
		db:       db,
		keychain: kc,
		router:   r,
		bus:      bus.New(),
	}

	s.routes()
	return s
}

func (s *Server) routes() {
	s.router.Get("/health", s.handleHealth)
	s.router.Get("/ws", s.handleWS)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // Allow all origins for local dev
	})
	if err != nil {
		// Log error to stdout for now
		return
	}
	defer c.Close(websocket.StatusInternalError, "internal error")

	// Subscribe to all events
	events, cancel := s.bus.Subscribe()
	defer cancel()

	ctx := r.Context()

	// Writer goroutine: Listen to bus, write to WS
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case evt := <-events:
				bytes, err := json.Marshal(evt)
				if err != nil {
					continue
				}
				c.Write(ctx, websocket.MessageText, bytes)
			}
		}
	}()

	// Reader loop: Keep connection alive and handle incoming messages if needed
	for {
		_, _, err := c.Read(ctx)
		if err != nil {
			break
		}
	}

	c.Close(websocket.StatusNormalClosure, "")
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.cfg.ListenAddr, s.router)
}
