//go:build !js || !wasm

package testutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CompileWASM compiles the Go application in appDir to a WASM binary at
// outputPath. It sets GOOS=js and GOARCH=wasm and runs `go build`.
// The appDir should contain a main package (or use a package path like
// "./src/app/").
func CompileWASM(appDir string, outputPath string) error {
	// Ensure the output directory exists
	outDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("testutil.CompileWASM: failed to create output directory: %w", err)
	}

	cmd := exec.Command("go", "build", "-o", outputPath, ".")
	cmd.Dir = appDir
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("testutil.CompileWASM: build failed: %w\nOutput: %s", err, string(output))
	}

	// Verify the output file was created
	info, err := os.Stat(outputPath)
	if err != nil {
		return fmt.Errorf("testutil.CompileWASM: output file not found after build: %w", err)
	}
	if info.Size() == 0 {
		return fmt.Errorf("testutil.CompileWASM: output file is empty")
	}

	return nil
}

// SetupWASMExecJS copies the Go runtime's wasm_exec.js file into the
// specified directory. This file is required to run Go WASM binaries in
// the browser.
func SetupWASMExecJS(dir string) error {
	srcPath, err := FindWASMExecJS()
	if err != nil {
		return fmt.Errorf("testutil.SetupWASMExecJS: %w", err)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("testutil.SetupWASMExecJS: failed to create directory: %w", err)
	}

	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("testutil.SetupWASMExecJS: failed to read %s: %w", srcPath, err)
	}

	dstPath := filepath.Join(dir, "wasm_exec.js")
	if err := os.WriteFile(dstPath, data, 0644); err != nil {
		return fmt.Errorf("testutil.SetupWASMExecJS: failed to write %s: %w", dstPath, err)
	}

	return nil
}

// FindWASMExecJS locates the wasm_exec.js file from the Go installation.
// It checks both the Go 1.21+ location (lib/wasm/) and the legacy location
// (misc/wasm/).
func FindWASMExecJS() (string, error) {
	goRoot, err := getGOROOT()
	if err != nil {
		return "", fmt.Errorf("failed to determine GOROOT: %w", err)
	}

	// Check both possible locations
	candidates := []string{
		filepath.Join(goRoot, "lib", "wasm", "wasm_exec.js"),  // Go 1.21+
		filepath.Join(goRoot, "misc", "wasm", "wasm_exec.js"), // Go < 1.21
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("wasm_exec.js not found in Go installation at %s (checked lib/wasm and misc/wasm)", goRoot)
}

// getGOROOT returns the GOROOT path, first checking the environment variable,
// then falling back to `go env GOROOT`.
func getGOROOT() (string, error) {
	goRoot := os.Getenv("GOROOT")
	if goRoot != "" {
		return goRoot, nil
	}

	cmd := exec.Command("go", "env", "GOROOT")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run 'go env GOROOT': %w", err)
	}

	goRoot = strings.TrimSpace(string(output))
	if goRoot == "" {
		return "", fmt.Errorf("GOROOT is empty")
	}

	return goRoot, nil
}
