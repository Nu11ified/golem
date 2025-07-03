package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Nu11ified/golem/internal/config"
)

// Builder handles building Golem applications
type Builder struct {
	config *config.Config
}

// NewBuilder creates a new Builder instance
func NewBuilder(config *config.Config) *Builder {
	return &Builder{
		config: config,
	}
}

// Build compiles the Golem application for production
func (b *Builder) Build() error {
	fmt.Println("📦 Preparing build...")

	// Clean build directory
	if err := b.cleanBuildDir(); err != nil {
		return fmt.Errorf("failed to clean build directory: %v", err)
	}

	// Parse .golem files and generate Go code
	fmt.Println("🔄 Parsing .golem files...")
	if err := b.parseGolemFiles(); err != nil {
		return fmt.Errorf("failed to parse .golem files: %v", err)
	}

	// Generate type definitions
	fmt.Println("🔧 Generating type definitions...")
	if err := b.generateTypes(); err != nil {
		return fmt.Errorf("failed to generate types: %v", err)
	}

	// Build WebAssembly binary
	fmt.Println("⚡ Building WebAssembly...")
	if err := b.buildWasm(); err != nil {
		return fmt.Errorf("failed to build WASM: %v", err)
	}

	// Build gRPC server
	fmt.Println("🔌 Building gRPC server...")
	if err := b.buildServer(); err != nil {
		return fmt.Errorf("failed to build server: %v", err)
	}

	// Generate static assets
	fmt.Println("📄 Generating static files...")
	if err := b.generateStaticFiles(); err != nil {
		return fmt.Errorf("failed to generate static files: %v", err)
	}

	return nil
}

func (b *Builder) cleanBuildDir() error {
	buildDir := b.config.Output
	if err := os.RemoveAll(buildDir); err != nil {
		return err
	}
	return os.MkdirAll(buildDir, 0755)
}

func (b *Builder) parseGolemFiles() error {
	// This would implement the .golem file parser
	// For now, we'll just copy files to build directory

	srcDir := "src"
	buildSrcDir := filepath.Join(b.config.Output, "src")

	return b.copyDir(srcDir, buildSrcDir)
}

func (b *Builder) generateTypes() error {
	// Generate TypeScript definitions for better IDE support
	typesDir := ".golem/types"
	if err := os.MkdirAll(typesDir, 0755); err != nil {
		return err
	}

	// This would analyze Go server functions and generate type definitions
	// For now, create a placeholder
	typeDef := `// Auto-generated type definitions for Golem
export interface ServerFunctions {
  Hello(name: string): Promise<string>;
  GetUserProfile(userID: number): Promise<UserProfile>;
}

export interface UserProfile {
  id: number;
  name: string;
  email: string;
}
`

	return os.WriteFile(filepath.Join(typesDir, "server.d.ts"), []byte(typeDef), 0644)
}

func (b *Builder) buildWasm() error {
	// Build the WebAssembly binary
	// Use absolute path for the output
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %v", err)
	}
	outputPath := filepath.Join(workingDir, b.config.Output, "app.wasm")

	cmd := exec.Command("go", "build", "-o", outputPath)
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	cmd.Dir = filepath.Join(b.config.Output, "src/app")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("WASM build failed: %v\nOutput: %s", err, output)
	}

	// Copy wasm_exec.js
	return b.copyWasmExec()
}

func (b *Builder) buildServer() error {
	// Build the gRPC server binary
	// Use absolute path for the output
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %v", err)
	}
	outputPath := filepath.Join(workingDir, b.config.Output, "server")

	cmd := exec.Command("go", "build", "-o", outputPath)
	cmd.Dir = filepath.Join(b.config.Output, "src/server")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("server build failed: %v\nOutput: %s", err, output)
	}

	return nil
}

func (b *Builder) generateStaticFiles() error {
	// Generate index.html
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>` + b.config.ProjectName + `</title>
    <style>
        body { font-family: system-ui, sans-serif; margin: 0; padding: 20px; }
        .app { max-width: 800px; margin: 0 auto; }
        .counter { margin: 20px 0; }
        .btn { padding: 8px 16px; margin: 4px; border: none; border-radius: 4px; cursor: pointer; }
        .btn-primary { background: #007bff; color: white; }
        .btn-secondary { background: #6c757d; color: white; }
        .btn-danger { background: #dc3545; color: white; }
        .btn-disabled { opacity: 0.6; cursor: not-allowed; }
    </style>
</head>
<body>
    <div id="app">Loading...</div>
    <script src="wasm_exec.js"></script>
    <script>
        const go = new Go();
        WebAssembly.instantiateStreaming(fetch("app.wasm"), go.importObject)
            .then((result) => {
                go.run(result.instance);
            });
    </script>
</body>
</html>`

	return os.WriteFile(filepath.Join(b.config.Output, "index.html"), []byte(html), 0644)
}

func (b *Builder) copyWasmExec() error {
	// Copy wasm_exec.js from Go installation
	goRoot := os.Getenv("GOROOT")
	if goRoot == "" {
		// Try to get GOROOT from go env
		cmd := exec.Command("go", "env", "GOROOT")
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get GOROOT: %v", err)
		}
		goRoot = strings.TrimSpace(string(output))
	}

	// Try both possible locations for wasm_exec.js
	possiblePaths := []string{
		filepath.Join(goRoot, "lib", "wasm", "wasm_exec.js"),  // Go 1.21+
		filepath.Join(goRoot, "misc", "wasm", "wasm_exec.js"), // Go < 1.21
	}

	var srcPath string
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			srcPath = path
			break
		}
	}

	if srcPath == "" {
		return fmt.Errorf("wasm_exec.js not found in Go installation at %s", goRoot)
	}

	dstPath := filepath.Join(b.config.Output, "wasm_exec.js")
	return b.copyFile(srcPath, dstPath)
}

func (b *Builder) copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return b.copyFile(path, dstPath)
	})
}

func (b *Builder) copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0644)
}
