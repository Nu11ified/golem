package serverfuncs

import (
	"os"
	"path/filepath"
	"strings"
)

type ServerFunction struct {
	Name     string
	Language string // "go" or "ts"
	Path     string
}

// DiscoverServerFunctions scans the user-app/server/go and user-app/server/ts directories
// and returns a list of available server functions.
func DiscoverServerFunctions(baseDir string) ([]ServerFunction, error) {
	var funcs []ServerFunction
	goDir := filepath.Join(baseDir, "server/go")
	tsDir := filepath.Join(baseDir, "server/ts")

	// Discover Go functions
	_ = filepath.Walk(goDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go") {
			name := strings.TrimSuffix(info.Name(), ".go")
			funcs = append(funcs, ServerFunction{
				Name:     name,
				Language: "go",
				Path:     path,
			})
		}
		return nil
	})

	// Discover TypeScript functions
	_ = filepath.Walk(tsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(info.Name(), ".ts") && !strings.HasSuffix(info.Name(), ".d.ts") {
			name := strings.TrimSuffix(info.Name(), ".ts")
			funcs = append(funcs, ServerFunction{
				Name:     name,
				Language: "ts",
				Path:     path,
			})
		}
		return nil
	})

	return funcs, nil
}
