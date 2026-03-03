package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestWASMCompilation(t *testing.T) {
	outDir := t.TempDir()
	outFile := filepath.Join(outDir, "app.wasm")

	cmd := exec.Command("go", "build", "-o", outFile, "./src/app/")
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")

	// Find repo root by walking up from cwd until we find go.mod
	wd, _ := os.Getwd()
	for !fileExists(filepath.Join(wd, "go.mod")) && wd != "/" {
		wd = filepath.Dir(wd)
	}
	cmd.Dir = wd

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("WASM compilation failed: %v\nOutput: %s", err, output)
	}

	info, err := os.Stat(outFile)
	if err != nil {
		t.Fatalf("WASM file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("WASM file is empty")
	}
	t.Logf("WASM binary size: %d bytes", info.Size())
}

func TestWasmExecJSExists(t *testing.T) {
	cmd := exec.Command("go", "env", "GOROOT")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("Could not get GOROOT: %v", err)
	}
	goRoot := strings.TrimSpace(string(out))

	paths := []string{
		filepath.Join(goRoot, "lib", "wasm", "wasm_exec.js"),
		filepath.Join(goRoot, "misc", "wasm", "wasm_exec.js"),
	}

	found := false
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			found = true
			t.Logf("Found wasm_exec.js at: %s", p)
			break
		}
	}
	if !found {
		t.Fatal("wasm_exec.js not found in Go installation")
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
