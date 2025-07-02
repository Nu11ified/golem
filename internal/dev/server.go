package dev

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Nu11ified/golem/internal/config"
	"nhooyr.io/websocket"
)

// Server represents the development server
type Server struct {
	config *config.Config
}

// NewServer creates a new development server
func NewServer(config *config.Config) *Server {
	return &Server{
		config: config,
	}
}

// Start starts the development server with hot reload
func (s *Server) Start() error {
	port := s.config.Dev.Port

	// Set up file watcher for hot reload
	if s.config.Dev.HotReload {
		go s.watchFiles()
	}

	// Set up HTTP handlers
	mux := http.NewServeMux()

	// Serve static files
	mux.Handle("/", s.createStaticHandler())

	// WebSocket endpoint for hot reload
	if s.config.Dev.HotReload {
		mux.HandleFunc("/ws", s.handleWebSocket)
	}

	fmt.Printf("🌟 Golem dev server running at http://localhost:%d\n", port)
	fmt.Println("📁 Serving files from:", s.config.Output)

	if s.config.Dev.HotReload {
		fmt.Println("🔥 Hot reload enabled")
	}

	return http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
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
    <div class="dev-banner">🔥 Development Mode - Hot Reload Enabled</div>
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
	fmt.Println("🔨 Building WebAssembly...")

	// Set environment variables for WebAssembly build
	env := append(os.Environ(),
		"GOOS=js",
		"GOARCH=wasm",
	)

	// Build command: go build -o app.wasm ./src/app/main.go
	wasmOutput := filepath.Join(devDir, "app.wasm")

	// Create a simple Go build command
	buildArgs := []string{
		"build",
		"-o", wasmOutput,
		"./src/app/main.go",
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

	fmt.Println("✅ WebAssembly build completed")
	return nil
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
