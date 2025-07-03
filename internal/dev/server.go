package dev

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Nu11ified/golem/internal/config"
	"github.com/Nu11ified/golem/internal/functions"
	"nhooyr.io/websocket"
)

// Server represents the development server
type Server struct {
	config   *config.Config
	registry *functions.Registry
}

// NewServer creates a new development server
func NewServer(config *config.Config) *Server {
	return &Server{
		config:   config,
		registry: functions.NewRegistry(),
	}
}

// Start starts the development server with hot reload and gRPC support
func (s *Server) Start() error {
	port := s.config.Dev.Port

	// Initialize function registry for development
	if err := s.initializeFunctionRegistry(); err != nil {
		log.Printf("Warning: Failed to initialize function registry: %v", err)
	}

	// Set up file watcher for hot reload
	if s.config.Dev.HotReload {
		go s.watchFiles()
	}

	// Start gRPC server in background for development
	go s.startDevGRPCServer()

	// Set up HTTP handlers
	mux := http.NewServeMux()

	// Serve static files
	mux.Handle("/", s.createStaticHandler())

	// API endpoint for function calls during development
	grpcServer := functions.NewGRPCServer(s.registry)
	mux.HandleFunc("/api/functions", grpcServer.HTTPHandler())

	// API root endpoint - show available endpoints
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/" {
			http.NotFound(w, r)
			return
		}

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusOK)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		// Get registered functions for display
		functions := s.registry.ListFunctions("")

		apiInfo := map[string]interface{}{
			"message": "Golem Development API",
			"version": "0.1.0",
			"endpoints": map[string]interface{}{
				"GET /api/":               "This endpoint - API information",
				"GET /api/functions/list": "List all registered server functions",
				"POST /api/functions":     "Call a server function",
			},
			"registered_functions": len(functions),
			"functions":            functions,
			"example_call": map[string]interface{}{
				"url":    "/api/functions",
				"method": "POST",
				"headers": map[string]string{
					"Content-Type": "application/json",
				},
				"body": map[string]interface{}{
					"serviceName":  "server",
					"functionName": "Hello",
					"args":         []interface{}{"World"},
				},
			},
			"grpc_server": map[string]interface{}{
				"port": s.config.Server.GRPC.Port,
				"url":  fmt.Sprintf("localhost:%d", s.config.Server.GRPC.Port),
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(apiInfo)
	})

	// List functions endpoint for development debugging
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

	// WebSocket endpoint for hot reload
	if s.config.Dev.HotReload {
		mux.HandleFunc("/ws", s.handleWebSocket)
	}

	fmt.Printf("🌟 Golem dev server running at http://localhost:%d\n", port)
	fmt.Println("📁 Serving files from:", s.config.Output)
	fmt.Printf("🔗 API endpoints available at: http://localhost:%d/api/\n", port)

	if s.config.Dev.HotReload {
		fmt.Println("🔥 Hot reload enabled")
	}

	return http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
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

	// Register user functions from the server package
	if err := s.registerUserFunctions(); err != nil {
		log.Printf("Warning: Failed to register user functions: %v", err)
	}

	log.Printf("🎯 Function registry initialized with %d functions", len(s.registry.ListFunctions("")))
	return nil
}

// registerUserFunctions registers functions from the user's server package
func (s *Server) registerUserFunctions() error {
	// First, try to build and import server packages to trigger their init() functions
	serverDir := s.config.Server.Functions
	if serverDir == "" {
		serverDir = "src/server"
	}

	// Build and import server packages (this will trigger init() functions)
	if err := s.registry.BuildAndImportServerPackages(serverDir); err != nil {
		log.Printf("Warning: Could not build server packages: %v", err)
	}

	// For development mode, register demo functions directly if they exist
	if err := s.registerDemoFunctions(); err != nil {
		log.Printf("Warning: Could not register demo functions: %v", err)
	}

	// Copy all functions from the global registry to this server's registry
	if err := s.registry.RegisterFromGlobal(); err != nil {
		return fmt.Errorf("failed to register functions from global registry: %w", err)
	}

	// Log registered functions
	registeredFunctions := s.registry.ListFunctions("")
	log.Printf("Successfully registered %d server functions:", len(registeredFunctions))
	for _, fn := range registeredFunctions {
		log.Printf("  - %s.%s", fn.ServiceName, fn.Name)
	}

	return nil
}

// registerDemoFunctions registers demo functions directly for development
func (s *Server) registerDemoFunctions() error {
	// Define demo functions that match the ones in the server package
	helloFunc := func(name string) string {
		return fmt.Sprintf("Hello, %s! This message is from the Go server.", name)
	}

	getUserProfileFunc := func(userID int) (map[string]interface{}, error) {
		if userID <= 0 {
			return nil, fmt.Errorf("invalid user ID: %d", userID)
		}
		return map[string]interface{}{
			"id":    userID,
			"name":  "John Doe",
			"email": "john@example.com",
			"role":  "admin",
		}, nil
	}

	calculateFunc := func(a, b float64, operation string) (float64, error) {
		switch operation {
		case "add":
			return a + b, nil
		case "subtract":
			return a - b, nil
		case "multiply":
			return a * b, nil
		case "divide":
			if b == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return a / b, nil
		default:
			return 0, fmt.Errorf("unknown operation: %s", operation)
		}
	}

	// Register the demo functions
	if err := s.registry.RegisterFunction("server", "Hello", helloFunc); err != nil {
		return fmt.Errorf("failed to register Hello function: %w", err)
	}

	if err := s.registry.RegisterFunction("server", "GetUserProfile", getUserProfileFunc); err != nil {
		return fmt.Errorf("failed to register GetUserProfile function: %w", err)
	}

	if err := s.registry.RegisterFunction("server", "Calculate", calculateFunc); err != nil {
		return fmt.Errorf("failed to register Calculate function: %w", err)
	}

	log.Printf("Registered demo functions for development")
	return nil
}

func (s *Server) startDevGRPCServer() {
	port := s.config.Server.GRPC.Port
	if port == 0 {
		port = 50051
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Printf("Warning: Failed to start dev gRPC server: %v", err)
		return
	}

	grpcServer := functions.CreateGRPCServer(s.registry)
	fmt.Printf("🔧 Dev gRPC server running at localhost:%d\n", port)

	if err := grpcServer.Serve(listener); err != nil {
		log.Printf("Dev gRPC server error: %v", err)
	}
}

func (s *Server) createStaticHandler() http.Handler {
	// Create a development version of the static files
	if err := s.generateDevFiles(); err != nil {
		log.Printf("Error generating dev files: %v", err)
	}

	// Serve from the development directory
	devDir := ".golem/dev"
	fs := http.FileServer(http.Dir(devDir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers for development
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Set proper MIME type for WASM files
		if filepath.Ext(r.URL.Path) == ".wasm" {
			w.Header().Set("Content-Type", "application/wasm")
		}

		fs.ServeHTTP(w, r)
	})
}

func (s *Server) generateDevFiles() error {
	// Ensure dev directory exists
	devDir := ".golem/dev"
	if err := os.MkdirAll(devDir, 0755); err != nil {
		return err
	}

	// Generate development HTML with hot reload
	html := s.generateDevHTML()
	htmlPath := filepath.Join(devDir, "index.html")
	if err := os.WriteFile(htmlPath, []byte(html), 0644); err != nil {
		return err
	}

	// Copy/build WASM for development
	return s.buildDevWasm()
}

func (s *Server) generateDevHTML() string {
	hotReloadScript := ""
	if s.config.Dev.HotReload {
		hotReloadScript = `
    <script>
        // Hot reload WebSocket connection
        const ws = new WebSocket('ws://localhost:` + fmt.Sprintf("%d", s.config.Dev.Port) + `/ws');
        ws.onmessage = function(event) {
            if (event.data === 'reload') {
                window.location.reload();
            }
        };
    </script>`
	}

	cacheBuster := fmt.Sprintf("%d", time.Now().UnixNano())

	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>` + s.config.ProjectName + ` - Development</title>
    <style>
        body { font-family: system-ui, sans-serif; margin: 0; padding: 20px; }
        .app { max-width: 800px; margin: 0 auto; }
        .counter { margin: 20px 0; }
        .btn { padding: 8px 16px; margin: 4px; border: none; border-radius: 4px; cursor: pointer; }
        .btn-primary { background: #007bff; color: white; }
        .btn-secondary { background: #6c757d; color: white; }
        .btn-danger { background: #dc3545; color: white; }
        .btn-disabled { opacity: 0.6; cursor: not-allowed; }
        .dev-banner { 
            background: #28a745; color: white; padding: 8px; text-align: center; 
            font-size: 12px; position: fixed; top: 0; left: 0; right: 0; z-index: 1000;
        }
        body { padding-top: 40px; }
    </style>
</head>
<body>
    <div class="dev-banner">🔥 Development Mode - Hot Reload Enabled | gRPC Server Active</div>
    <div id="app">Loading Golem app...</div>
    <script src="wasm_exec.js?` + cacheBuster + `"></script>
    <script>
        const go = new Go();
        // Shim for older wasm_exec.js files with newer Go compilers.
        if (go.importObject.go && !go.importObject.gojs) {
            console.log("Shimming go->gojs import object.");
            go.importObject.gojs = go.importObject.go;
            delete go.importObject.go;
        }

        console.log("Attempting to instantiate wasm with import object:", go.importObject);

        const wasmModule = fetch("app.wasm?` + cacheBuster + `");
        const instantiateWasm = async () => {
            try {
                let instance;
                if (WebAssembly.instantiateStreaming) {
                    instance = (await WebAssembly.instantiateStreaming(wasmModule, go.importObject)).instance;
                } else {
                    const response = await wasmModule;
                    const bytes = await response.arrayBuffer();
                    instance = (await WebAssembly.instantiate(bytes, go.importObject)).instance;
                }
                go.run(instance);
            } catch (err) {
                document.getElementById('app').innerHTML =
                    '<h1>❌ Error loading WebAssembly</h1>' +
                    '<h2>See browser developer console for details.</h2>' +
                    '<pre>' + err.toString() + '</pre>';
                console.error('WASM Instantiation Error:', err);
                console.error('Import object passed to instantiate:', go.importObject);
            }
        };
        instantiateWasm();
    </script>` + hotReloadScript + `
</body>
</html>`
}

func (s *Server) buildDevWasm() error {
	// First, ensure server packages are discovered and import file is generated
	serverDir := s.config.Server.Functions
	if serverDir == "" {
		serverDir = "src/server"
	}

	// Generate server imports file
	if err := s.registry.BuildAndImportServerPackages(serverDir); err != nil {
		log.Printf("Warning: Could not generate server imports: %v", err)
	}

	// Find wasm_exec.js from Go installation.
	var wasmExecSrc string
	goRootCmd := exec.Command("go", "env", "GOROOT")
	goRootBytes, err := goRootCmd.Output()
	if err != nil {
		log.Println("WARNING: Could not run `go env GOROOT` to find wasm_exec.js. Will use a potentially incomplete fallback. Please ensure `go` is in your PATH.")
	} else {
		goRoot := strings.TrimSpace(string(goRootBytes))
		// Check potential locations for wasm_exec.js
		potentialPaths := []string{
			filepath.Join(goRoot, "misc", "wasm", "wasm_exec.js"), // Go < 1.21
			filepath.Join(goRoot, "lib", "wasm", "wasm_exec.js"),  // Go >= 1.21
		}
		for _, p := range potentialPaths {
			if _, err := os.Stat(p); err == nil {
				log.Printf("Found wasm_exec.js at: %s", p)
				wasmExecSrc = p
				break
			}
		}
	}

	devDir := ".golem/dev"
	wasmExecDest := filepath.Join(devDir, "wasm_exec.js")

	// If we found the official file, use it. Otherwise, use the fallback.
	if wasmExecSrc != "" {
		// Copy the official wasm_exec.js
		wasmExecData, err := os.ReadFile(wasmExecSrc)
		if err != nil {
			return fmt.Errorf("failed to read wasm_exec.js from %s: %v", wasmExecSrc, err)
		}
		if err := os.WriteFile(wasmExecDest, wasmExecData, 0644); err != nil {
			return fmt.Errorf("failed to copy wasm_exec.js: %v", err)
		}
	} else {
		// Create a minimal wasm_exec.js fallback
		wasmExecContent := `// Minimal wasm_exec.js for Golem development
class Go {
	constructor() {
		this.importObject = {
			gojs: {
				"debug": console.log,
				"runtime.wasmExit": () => {},
				"runtime.wasmWrite": (fd, p) => {
					if (fd === 1) {
						console.log(new TextDecoder().decode(p));
					} else {
						console.error(new TextDecoder().decode(p));
					}
				},
				"runtime.resetMemoryDataView": () => {},
				"runtime.nanotime1": () => Date.now() * 1000000,
				"runtime.walltime": () => {
					const msec = Date.now();
					return [Math.floor(msec / 1000), (msec % 1000) * 1000000];
				},
				"runtime.scheduleTimeoutEvent": () => {},
				"runtime.clearTimeoutEvent": () => {},
				"runtime.getRandomData": (buf) => {
					if (typeof crypto !== 'undefined' && crypto.getRandomValues) {
						crypto.getRandomValues(buf);
					} else {
						for (let i = 0; i < buf.length; i++) {
							buf[i] = Math.floor(Math.random() * 256);
						}
					}
				},
				"syscall/js.finalizeRef": () => {},
				"syscall/js.valueGet": () => {},
				"syscall/js.valueSet": () => {},
				"syscall/js.valueDelete": () => {},
				"syscall/js.valueIndex": () => {},
				"syscall/js.valueSetIndex": () => {},
				"syscall/js.valueCall": () => {},
				"syscall/js.valueInvoke": () => {},
				"syscall/js.valueNew": () => {},
				"syscall/js.valueLength": () => {},
				"syscall/js.valuePrepareString": () => {},
				"syscall/js.valueLoadString": () => {},
				"syscall/js.valueInstanceOf": () => {},
				"syscall/js.copyBytesToGo": () => {},
				"syscall/js.copyBytesToJS": () => {},
			}
		};
		this.memory = null;
	}
	
	run(instance) {
		this.memory = instance.exports.mem;
		instance.exports.run();
	}
}

if (typeof module !== 'undefined' && module.exports) {
	module.exports = { Go };
} else {
	window.Go = Go;
}
`
		if err := os.WriteFile(wasmExecDest, []byte(wasmExecContent), 0644); err != nil {
			return fmt.Errorf("failed to create wasm_exec.js: %v", err)
		}
	}

	// Build the WebAssembly file
	fmt.Println("🔨 Building WebAssembly with server functions...")

	// Set environment variables for WebAssembly build
	env := append(os.Environ(),
		"GOOS=js",
		"GOARCH=wasm",
	)

	// Build command: go build -o app.wasm ./src/app/main.go
	wasmOutput := filepath.Join(devDir, "app.wasm")

	// Create a temporary main.go that imports server packages
	tempMainFile := filepath.Join(devDir, "main.go")
	if err := s.createWasmMainFile(tempMainFile); err != nil {
		return fmt.Errorf("failed to create WASM main file: %v", err)
	}

	// Build the WASM file from the temporary main
	buildArgs := []string{
		"build",
		"-o", wasmOutput,
		tempMainFile,
	}

	// Execute go build
	cmd := exec.Command("go", buildArgs...)
	cmd.Dir = "."
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("WebAssembly build failed: %v", err)
	}

	fmt.Println("✅ WebAssembly build completed with server functions")
	return nil
}

// createWasmMainFile creates a main.go file that imports both app and server packages
func (s *Server) createWasmMainFile(mainFile string) error {
	// Read the original main.go to get its content
	originalMain, err := os.ReadFile("src/app/main.go")
	if err != nil {
		return fmt.Errorf("failed to read original main.go: %v", err)
	}

	// Get module name for proper imports
	moduleName, err := functions.GetModuleName()
	if err != nil {
		return fmt.Errorf("failed to get module name: %v", err)
	}

	// Create a new main.go that imports server packages
	content := fmt.Sprintf(`//go:build js && wasm

// Auto-generated main.go for WASM build with server functions
package main

import (
	_ "%s/src/server" // Import server package to trigger function registration
)

// Include the original main.go content below
`, moduleName)

	// Append the original main.go content, but remove its package declaration
	originalContent := string(originalMain)
	lines := strings.Split(originalContent, "\n")
	var filteredLines []string

	for i, line := range lines {
		// Skip the package main line from original file
		if i == 0 && strings.HasPrefix(strings.TrimSpace(line), "package main") {
			continue
		}
		filteredLines = append(filteredLines, line)
	}

	content += strings.Join(filteredLines, "\n")

	// Write the combined main.go
	return os.WriteFile(mainFile, []byte(content), 0644)
}

func (s *Server) watchFiles() {
	// File watcher implementation for hot reload
	// This would watch the files specified in config.Dev.Watch
	log.Println("🔍 File watcher started")

	// Placeholder - would implement actual file watching
	// using fsnotify or similar
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// This is a basic WebSocket handler to prevent connection errors in the browser.
	// Full hot-reload logic is not implemented yet.
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Printf("could not upgrade to websocket: %v", err)
		return
	}
	defer c.Close(websocket.StatusInternalError, "internal error")

	log.Println("WebSocket client connected.")

	// Keep the connection open but do nothing.
	// This prevents the connection from being immediately closed and causing errors.
	for {
		_, _, err := c.Read(r.Context())
		if err != nil {
			break
		}
	}
}
