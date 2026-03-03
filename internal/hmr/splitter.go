package hmr

import (
	"os"
	"path/filepath"
	"strings"
)

// AffectedResult describes which modules are affected by a file change.
type AffectedResult struct {
	// IsShell indicates whether the change affects the shell (app-level) module,
	// which requires a full page reload.
	IsShell bool

	// PagePaths contains the paths of affected page modules.
	PagePaths []string

	// ModulePath is the primary module path affected by the change.
	ModulePath string
}

// ModuleSplitter analyzes a project directory to understand its module structure
// and determine which modules are affected by file changes.
type ModuleSplitter struct {
	// rootDir is the project root directory.
	rootDir string

	// shellDirs are directories considered part of the shell/app module.
	shellDirs []string

	// pageDirs are directories considered page modules.
	pageDirs []string
}

// NewModuleSplitter creates a new ModuleSplitter for the given root directory.
func NewModuleSplitter(rootDir string) *ModuleSplitter {
	return &ModuleSplitter{
		rootDir:   rootDir,
		shellDirs: []string{"src/app", "src/components"},
		pageDirs:  []string{"src/pages"},
	}
}

// Analyze scans the project directory to discover the module structure.
func (ms *ModuleSplitter) Analyze(dir string) error {
	ms.rootDir = dir

	// Discover page directories
	pagesDir := filepath.Join(dir, "src", "pages")
	if _, err := os.Stat(pagesDir); err == nil {
		entries, err := os.ReadDir(pagesDir)
		if err != nil {
			return err
		}
		ms.pageDirs = nil
		for _, entry := range entries {
			if entry.IsDir() {
				ms.pageDirs = append(ms.pageDirs, filepath.Join("src", "pages", entry.Name()))
			}
		}
		// If no subdirectories, treat the pages dir itself as a page dir
		if len(ms.pageDirs) == 0 {
			ms.pageDirs = []string{"src/pages"}
		}
	}

	return nil
}

// DetermineAffectedModules determines which modules are affected by a change
// to the given file path.
func (ms *ModuleSplitter) DetermineAffectedModules(changedFile string) AffectedResult {
	// Normalize the path relative to root
	relPath := changedFile
	if filepath.IsAbs(changedFile) && ms.rootDir != "" {
		rel, err := filepath.Rel(ms.rootDir, changedFile)
		if err == nil {
			relPath = rel
		}
	}

	// Normalize separators
	relPath = filepath.ToSlash(relPath)

	// Check if the file is in a page directory.
	// For page directories like "src/pages", we identify the specific page module
	// by extracting the first subdirectory under the pages root.
	// For example, "src/pages/home/index.go" -> page module is "src/pages/home".
	for _, pageDir := range ms.pageDirs {
		normalizedPageDir := filepath.ToSlash(pageDir)
		if strings.HasPrefix(relPath, normalizedPageDir+"/") || relPath == normalizedPageDir {
			modulePath := ms.extractPageModule(normalizedPageDir, relPath)
			return AffectedResult{
				IsShell:    false,
				PagePaths:  []string{modulePath},
				ModulePath: modulePath,
			}
		}
	}

	// Check if the file is in a shell directory
	for _, shellDir := range ms.shellDirs {
		normalizedShellDir := filepath.ToSlash(shellDir)
		if strings.HasPrefix(relPath, normalizedShellDir+"/") || relPath == normalizedShellDir {
			return AffectedResult{
				IsShell:    true,
				PagePaths:  nil,
				ModulePath: normalizedShellDir,
			}
		}
	}

	// Default: treat unknown files as shell changes (full reload)
	return AffectedResult{
		IsShell:    true,
		PagePaths:  nil,
		ModulePath: relPath,
	}
}

// extractPageModule returns the page module directory for a given file path.
// If the pageDir is "src/pages" and the file is "src/pages/home/index.go",
// the page module is "src/pages/home". If the file is directly in "src/pages",
// the module is "src/pages" itself.
func (ms *ModuleSplitter) extractPageModule(pageDir, filePath string) string {
	// Get the path relative to the page directory
	suffix := strings.TrimPrefix(filePath, pageDir+"/")
	if suffix == filePath {
		// No suffix means the file IS the page dir
		return pageDir
	}

	// Extract the first path component (the page subdirectory name)
	parts := strings.SplitN(suffix, "/", 2)
	if len(parts) > 0 && parts[0] != "" {
		return pageDir + "/" + parts[0]
	}

	return pageDir
}
