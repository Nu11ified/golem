package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Nu11ified/golem/internal/config"
)

// Server represents the production server
type Server struct {
	config     *config.Config
	httpServer *http.Server
}

// NewServer creates a new production server
func NewServer(config *config.Config) *Server {
	return &Server{
		config: config,
	}
}

// Start starts the production server
func (s *Server) Start() error {
	// For now, just start HTTP server for static files
	// gRPC functionality will be added later
	return s.startHTTPServer()
}

func (s *Server) startHTTPServer() error {
	mux := http.NewServeMux()

	// Serve static files from build directory
	fs := http.FileServer(http.Dir(s.config.Output))
	mux.Handle("/", fs)

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	port := 8080 // Default HTTP port for production
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	fmt.Printf("🚀 Production server running at http://localhost:%d\n", port)
	fmt.Printf("📁 Serving static files from: %s\n", s.config.Output)

	return s.httpServer.ListenAndServe()
}

// gRPC functionality will be implemented later
// func (s *Server) startGRPCServer() error {
//     // Implementation will be added when gRPC dependencies are available
//     return nil
// }

// Stop gracefully stops the server
func (s *Server) Stop() error {
	if s.httpServer != nil {
		if err := s.httpServer.Close(); err != nil {
			log.Printf("Error closing HTTP server: %v", err)
		}
	}

	return nil
}
