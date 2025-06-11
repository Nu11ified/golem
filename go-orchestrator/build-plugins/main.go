package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	goDir := filepath.Join("..", "user-app", "server", "go")
	files, err := os.ReadDir(goDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read %s: %v\n", goDir, err)
		os.Exit(1)
	}
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".go") || strings.HasSuffix(f.Name(), "_test.go") {
			continue
		}
		src := filepath.Join(goDir, f.Name())
		out := filepath.Join(goDir, strings.TrimSuffix(f.Name(), ".go")+".so")
		fmt.Printf("Building %s -> %s\n", src, out)
		cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", out, src)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to build %s: %v\n", src, err)
			os.Exit(1)
		}
	}
}
