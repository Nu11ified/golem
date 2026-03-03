//go:build !js || !wasm

package testutil

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// TestServer wraps an HTTP server suitable for end-to-end testing of Golem
// applications. It serves static files from the app's build directory and
// provides a health check endpoint. The server listens on a random available
// port so tests can run in parallel without port conflicts.
type TestServer struct {
	server   *http.Server
	listener net.Listener
	dir      string // app root directory
	buildDir string // directory containing built assets
}

// StartTestServer starts a test server that serves static files from the
// build directory within dir. It automatically picks a free port.
// The server exposes:
//   - / — static file server for the build directory
//   - /health — returns 200 OK
//
// The caller must call Stop when done.
func StartTestServer(dir string) (*TestServer, error) {
	buildDir := filepath.Join(dir, "build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return nil, fmt.Errorf("testutil.StartTestServer: failed to create build directory: %w", err)
	}

	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Static file server for built assets
	fs := http.FileServer(http.Dir(buildDir))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set correct MIME type for WASM files
		if filepath.Ext(r.URL.Path) == ".wasm" {
			w.Header().Set("Content-Type", "application/wasm")
		}
		fs.ServeHTTP(w, r)
	}))

	// Bind to a random available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("testutil.StartTestServer: failed to listen: %w", err)
	}

	srv := &http.Server{
		Handler: mux,
	}

	ts := &TestServer{
		server:   srv,
		listener: listener,
		dir:      dir,
		buildDir: buildDir,
	}

	// Start serving in the background
	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "testutil.TestServer: serve error: %v\n", err)
		}
	}()

	return ts, nil
}

// URL returns the base URL of the test server (e.g., "http://127.0.0.1:12345").
func (s *TestServer) URL() string {
	return "http://" + s.listener.Addr().String()
}

// Port returns the port number the test server is listening on.
func (s *TestServer) Port() int {
	return s.listener.Addr().(*net.TCPAddr).Port
}

// BuildDir returns the path to the build directory being served.
func (s *TestServer) BuildDir() string {
	return s.buildDir
}

// WaitReady polls the server's /health endpoint until it responds with
// HTTP 200 or the timeout expires. This is useful to ensure the server
// is fully ready before running tests against it.
func (s *TestServer) WaitReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	healthURL := s.URL() + "/health"

	client := &http.Client{Timeout: 1 * time.Second}

	for time.Now().Before(deadline) {
		resp, err := client.Get(healthURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(10 * time.Millisecond)
	}

	return fmt.Errorf("testutil.WaitReady: server at %s did not become ready within %v", s.URL(), timeout)
}

// Stop gracefully shuts down the test server with a 5-second timeout.
func (s *TestServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = s.server.Shutdown(ctx)
}

// StartTestServerWithHandler starts a test server using a custom HTTP handler
// instead of the default static file server. This is useful for testing
// specific server behaviors.
func StartTestServerWithHandler(handler http.Handler) (*TestServer, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("testutil.StartTestServerWithHandler: failed to listen: %w", err)
	}

	srv := &http.Server{
		Handler: handler,
	}

	ts := &TestServer{
		server:   srv,
		listener: listener,
	}

	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "testutil.TestServer: serve error: %v\n", err)
		}
	}()

	return ts, nil
}
