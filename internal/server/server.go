package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/Nu11ified/golem/internal/config"
	"github.com/Nu11ified/golem/internal/functions"
	"google.golang.org/grpc"
)

// Server represents the production server
type Server struct {
	config     *config.Config
	httpServer *http.Server
	grpcServer *grpc.Server
	registry   *functions.Registry
}

// NewServer creates a new production server
func NewServer(config *config.Config) *Server {
	return &Server{
		config:   config,
		registry: functions.NewRegistry(),
	}
}

// Start starts the production server (both HTTP and gRPC)
func (s *Server) Start() error {
	// Initialize the function registry
	if err := s.initializeFunctionRegistry(); err != nil {
		return fmt.Errorf("failed to initialize function registry: %w", err)
	}

	// Start both servers concurrently
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	// Start HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.startHTTPServer(); err != nil {
			errChan <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	// Start gRPC server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.startGRPCServer(); err != nil {
			errChan <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	// Wait for first error or completion
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Return first error encountered
	return <-errChan
}

func (s *Server) initializeFunctionRegistry() error {
	// Discover functions from the server directory
	serverDir := s.config.Server.Functions
	if serverDir == "" {
		serverDir = "src/server"
	}

	if err := s.registry.DiscoverFunctions(serverDir); err != nil {
		log.Printf("Warning: Failed to discover functions from %s: %v", serverDir, err)
	}

	// Initialize user function registry
	if err := s.registerUserFunctions(); err != nil {
		log.Printf("Warning: Failed to initialize user functions: %v", err)
	}

	return nil
}

func (s *Server) registerUserFunctions() error {
	// This would load actual user-defined functions from their server directory
	// For now, just log that we're ready to accept function registrations
	log.Printf("Function registry ready for user function registration")
	return nil
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

	// API endpoint for function calls (HTTP bridge to gRPC)
	grpcServer := functions.NewGRPCServer(s.registry)
	mux.HandleFunc("/api/functions", grpcServer.HTTPHandler())

	// List functions endpoint
	mux.HandleFunc("/api/functions/list", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusOK)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		functions := s.registry.ListFunctions("")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"functions": functions,
		})
	})

	port := 8080 // Default HTTP port for production
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	fmt.Printf("🚀 Production HTTP server running at http://localhost:%d\n", port)
	fmt.Printf("📁 Serving static files from: %s\n", s.config.Output)
	fmt.Printf("🔗 API endpoints available at: http://localhost:%d/api/\n", port)

	return s.httpServer.ListenAndServe()
}

func (s *Server) startGRPCServer() error {
	port := s.config.Server.GRPC.Port
	if port == 0 {
		port = 50051 // Default gRPC port
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to create gRPC listener: %w", err)
	}

	s.grpcServer = functions.CreateGRPCServer(s.registry)

	fmt.Printf("🔧 gRPC server running at localhost:%d\n", port)
	fmt.Printf("🎯 Available functions: %d\n", len(s.registry.ListFunctions("")))

	return s.grpcServer.Serve(listener)
}

// Stop gracefully stops both servers
func (s *Server) Stop() error {
	var errors []error

	if s.httpServer != nil {
		if err := s.httpServer.Close(); err != nil {
			errors = append(errors, fmt.Errorf("HTTP server close error: %w", err))
		}
	}

	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}

	if len(errors) > 0 {
		return fmt.Errorf("server shutdown errors: %v", errors)
	}

	return nil
}
