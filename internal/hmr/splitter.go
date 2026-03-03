package hmr

import (
	"fmt"
	"go/format"
	"path/filepath"
	"strings"

	"github.com/Nu11ified/golem/internal/codegen"
)

// ModuleInfo represents a compilable WASM module
type ModuleInfo struct {
	Name        string   // e.g., "shell", "page_about", "page_blog_slug"
	EntryFile   string   // main Go file for this module
	OutputPath  string   // where to write the WASM binary
	IsShell     bool     // true for the shell module
	RoutePath   string   // URL path this page handles (empty for shell)
	SourceFiles []string // Go files that belong to this module
}

// ModuleSplitter analyzes the app structure and determines module boundaries
type ModuleSplitter struct {
	appDir     string
	outputDir  string
	moduleName string
}

// NewModuleSplitter creates a new ModuleSplitter.
// appDir is the path to the app source directory (e.g., "src/app").
// outputDir is where WASM binaries will be written.
// moduleName is the Go module name (e.g., "github.com/example/myapp").
func NewModuleSplitter(appDir, outputDir, moduleName string) *ModuleSplitter {
	return &ModuleSplitter{
		appDir:     appDir,
		outputDir:  outputDir,
		moduleName: moduleName,
	}
}

// Analyze determines what modules to build by scanning the app directory
// for page.go files and constructing module boundaries.
// It returns a slice of ModuleInfo: one shell module plus one module per page.
func (s *ModuleSplitter) Analyze() ([]ModuleInfo, error) {
	root, err := codegen.ScanRoutes(s.appDir)
	if err != nil {
		return nil, fmt.Errorf("failed to scan routes: %w", err)
	}

	var modules []ModuleInfo

	// Collect all page modules and shell source files
	var shellSourceFiles []string
	s.collectPages(root, &modules, &shellSourceFiles)

	// Build the shell module (always first)
	shell := ModuleInfo{
		Name:        "shell",
		EntryFile:   filepath.Join(s.appDir, "main.go"),
		OutputPath:  filepath.Join(s.outputDir, "shell.wasm"),
		IsShell:     true,
		RoutePath:   "",
		SourceFiles: shellSourceFiles,
	}

	// Prepend shell as the first module
	result := make([]ModuleInfo, 0, len(modules)+1)
	result = append(result, shell)
	result = append(result, modules...)

	return result, nil
}

// collectPages recursively walks the route tree collecting page modules
// and shell source files.
func (s *ModuleSplitter) collectPages(route *codegen.ScannedRoute, pageModules *[]ModuleInfo, shellFiles *[]string) {
	// If this route has a page.go, create a page module for it
	if route.PageFile != "" {
		moduleName := s.pageModuleName(route)
		mod := ModuleInfo{
			Name:        moduleName,
			EntryFile:   route.PageFile,
			OutputPath:  filepath.Join(s.outputDir, moduleName+".wasm"),
			IsShell:     false,
			RoutePath:   route.Path,
			SourceFiles: []string{route.PageFile},
		}
		*pageModules = append(*pageModules, mod)
	}

	// Non-page files in this directory belong to the shell
	for _, f := range s.shellFilesForRoute(route) {
		*shellFiles = append(*shellFiles, f)
	}

	// Recurse into children
	for _, child := range route.Children {
		s.collectPages(child, pageModules, shellFiles)
	}

	// Recurse into parallel slots
	for _, slot := range route.ParallelSlots {
		s.collectPages(slot, pageModules, shellFiles)
	}
}

// shellFilesForRoute returns the non-page special files for a route
// that belong to the shell module.
func (s *ModuleSplitter) shellFilesForRoute(route *codegen.ScannedRoute) []string {
	var files []string
	if route.LayoutFile != "" {
		files = append(files, route.LayoutFile)
	}
	if route.ErrorFile != "" {
		files = append(files, route.ErrorFile)
	}
	if route.LoadingFile != "" {
		files = append(files, route.LoadingFile)
	}
	if route.NotFoundFile != "" {
		files = append(files, route.NotFoundFile)
	}
	if route.TemplateFile != "" {
		files = append(files, route.TemplateFile)
	}
	return files
}

// pageModuleName generates a module name from the route.
// Root page.go -> "page_root"
// src/app/about/page.go -> "page_about"
// src/app/blog/[slug]/page.go -> "page_blog_slug"
func (s *ModuleSplitter) pageModuleName(route *codegen.ScannedRoute) string {
	routePath := route.Path
	if routePath == "/" {
		return "page_root"
	}

	// Remove leading slash
	routePath = strings.TrimPrefix(routePath, "/")

	// Replace path separators with underscores
	routePath = strings.ReplaceAll(routePath, "/", "_")

	// Remove special route characters (:, *)
	routePath = strings.ReplaceAll(routePath, ":", "")
	routePath = strings.ReplaceAll(routePath, "*", "")

	// Clean up any double underscores
	for strings.Contains(routePath, "__") {
		routePath = strings.ReplaceAll(routePath, "__", "_")
	}

	// Trim trailing underscores
	routePath = strings.TrimRight(routePath, "_")

	return "page_" + routePath
}

// GenerateShellEntry generates the shell module's main.go that contains
// the router, layout system, and module loader.
func (s *ModuleSplitter) GenerateShellEntry(modules []ModuleInfo) (string, error) {
	var pageModules []ModuleInfo
	for _, m := range modules {
		if !m.IsShell {
			pageModules = append(pageModules, m)
		}
	}

	var routeEntries strings.Builder
	for _, m := range pageModules {
		routeEntries.WriteString(fmt.Sprintf("\t\t%q: %q,\n", m.RoutePath, m.Name+".wasm"))
	}

	src := fmt.Sprintf(`//go:build js && wasm

package main

// Shell module: persistent WASM module containing router, state management,
// layout system, and CSS engine. Page modules are loaded dynamically.

// pageModules maps route paths to their WASM module filenames.
var pageModules = map[string]string{
%s}

func main() {
	// Shell stays running; page modules are loaded on navigation.
	select {}
}
`, routeEntries.String())

	formatted, err := format.Source([]byte(src))
	if err != nil {
		return "", fmt.Errorf("failed to format shell entry: %w", err)
	}

	return string(formatted), nil
}

// GeneratePageEntry generates a page module's main.go with //go:wasmexport.
// The generated file imports the page's package and exports a RenderPage function.
func (s *ModuleSplitter) GeneratePageEntry(module ModuleInfo) (string, error) {
	// Determine the import path for the page package.
	// The page.go file is in a directory; we need the Go import path for that directory.
	pageDir := filepath.Dir(module.EntryFile)

	// Compute relative path from appDir to the page directory
	relPath, err := filepath.Rel(s.appDir, pageDir)
	if err != nil {
		return "", fmt.Errorf("failed to compute relative path: %w", err)
	}

	// Build the import path
	var importPath string
	if relPath == "." {
		// Root page - import the app directory itself
		importPath = s.moduleName + "/src/app"
	} else {
		importPath = s.moduleName + "/src/app/" + filepath.ToSlash(relPath)
	}

	// Determine the package alias from the last directory component
	pkgAlias := filepath.Base(pageDir)
	// Clean up special characters from directory names like [slug]
	pkgAlias = strings.NewReplacer(
		"[", "",
		"]", "",
		".", "",
		"-", "_",
	).Replace(pkgAlias)

	src := fmt.Sprintf(`//go:build js && wasm

package main

import (
	page %q
)

// RenderPage is exported to the WASM host for HMR page swapping.
//
//go:wasmexport RenderPage
func RenderPage() {
	_ = page.Page
}

func main() {}
`, importPath)

	formatted, err := format.Source([]byte(src))
	if err != nil {
		return "", fmt.Errorf("failed to format page entry for %s: %w", module.Name, err)
	}

	return string(formatted), nil
}

// DetermineAffectedModules returns which modules need rebuilding given a set
// of changed files. If a page.go changed, only that page module is affected.
// If a layout, component, or other shared file changed, the shell module is affected.
func (s *ModuleSplitter) DetermineAffectedModules(changedFiles []string, allModules []ModuleInfo) []ModuleInfo {
	affectedSet := make(map[string]bool)
	var affected []ModuleInfo

	for _, changedFile := range changedFiles {
		absChanged, _ := filepath.Abs(changedFile)
		found := false

		// Check if the changed file belongs to a page module
		for _, m := range allModules {
			if m.IsShell {
				continue
			}
			for _, src := range m.SourceFiles {
				absSrc, _ := filepath.Abs(src)
				if absSrc == absChanged {
					if !affectedSet[m.Name] {
						affectedSet[m.Name] = true
						affected = append(affected, m)
					}
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		// If the changed file is not a page source, it affects the shell
		if !found {
			if !affectedSet["shell"] {
				affectedSet["shell"] = true
				for _, m := range allModules {
					if m.IsShell {
						affected = append(affected, m)
						break
					}
				}
			}
		}
	}

	return affected
}
