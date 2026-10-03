package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Aksoom-Hussain/Hospit-Open-Booking-Integration-Platform/internal/orchestrator"
	"github.com/Aksoom-Hussain/Hospit-Open-Booking-Integration-Platform/pkg/connector"
)

// Server wraps http.Server and handles routing
type Server struct {
	httpServer *http.Server
	handler    *Handler
}

// NewServer configures and returns an API server instance
func NewServer(port int, orch *orchestrator.Orchestrator, reg *connector.Registry) *Server {
	h := NewHandler(orch, reg)
	mux := http.NewServeMux()

	// Core Endpoints
	mux.HandleFunc("/v1/health", h.HandleHealth)
	mux.HandleFunc("/v1/suppliers", h.HandleListSuppliers)
	mux.HandleFunc("/v1/hotels/search", h.HandleSearch)
	mux.HandleFunc("/v1/hotels/quote", h.HandleQuote)
	mux.HandleFunc("/v1/bookings", h.HandleBooking)
	mux.HandleFunc("/v1/bookings/cancel", h.HandleCancel)
	mux.HandleFunc("/v1/ari/push", h.HandlePushARI)

	// Middleware pipeline
	rootHandler := loggingMiddleware(corsMiddleware(mux))

	return &Server{
		handler: h,
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      rootHandler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

// Start runs the HTTP server listening on the configured port
func (s *Server) Start() error {
	log.Printf("[Hospit Engine] Server listening on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully terminates the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// corsMiddleware sets standard CORS headers for local SaaS and public API consumers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Tenant-Key, X-Hospit-API-Key")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs incoming HTTP requests with timing
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s %s (%v)", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
	})
}
