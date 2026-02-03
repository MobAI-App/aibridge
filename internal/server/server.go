package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/MobAI-App/aibridge/internal/bridge"
)

type Server struct {
	httpServer *http.Server
	handlers   *Handlers
	verbose    bool
}

func New(b *bridge.Bridge, host string, port int, verbose bool) *Server {
	handlers := NewHandlers(b)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handlers.Health)
	mux.HandleFunc("GET /status", handlers.Status)
	mux.HandleFunc("POST /inject", handlers.Inject)
	mux.HandleFunc("DELETE /queue", handlers.QueueClear)
	mux.HandleFunc("OPTIONS /", handlePreflight)
	mux.HandleFunc("OPTIONS /health", handlePreflight)
	mux.HandleFunc("OPTIONS /status", handlePreflight)
	mux.HandleFunc("OPTIONS /inject", handlePreflight)
	mux.HandleFunc("OPTIONS /queue", handlePreflight)

	addr := fmt.Sprintf("%s:%d", host, port)

	return &Server{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: cors(mux),
		},
		handlers: handlers,
		verbose:  verbose,
	}
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		next.ServeHTTP(w, r)
	})
}

func handlePreflight(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) Start() error {
	if s.verbose {
		log.Printf("HTTP server starting on %s", s.httpServer.Addr)
	}
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) GracefulShutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = s.Shutdown(ctx)
}
