//go:build !js || !wasm

package staticgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/internal/ssr"
	"github.com/Nu11ified/golem/router"
)

// Generator pre-renders pages to static HTML at build time.
type Generator struct {
	renderer *ssr.SSRRenderer
	config   GeneratorConfig
}

// NewGenerator creates a static page generator.
func NewGenerator(r *router.Router, opts dom.DocumentOptions, config GeneratorConfig) *Generator {
	return &Generator{
		renderer: ssr.NewSSRRenderer(r, opts),
		config:   config,
	}
}

// Result represents the outcome of generating a single page.
type Result struct {
	Path    string
	OutFile string
	HTML    string
	Error   error
	Config  PageConfig
}

// Generate pre-renders all configured pages to static HTML.
func (g *Generator) Generate(pages []PageConfig) ([]Result, error) {
	if err := os.MkdirAll(g.config.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	concurrency := g.config.Concurrency
	if concurrency <= 0 {
		concurrency = 4
	}

	results := make([]Result, len(pages))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i, page := range pages {
		wg.Add(1)
		go func(idx int, pc PageConfig) {
			defer wg.Done()
			sem <- struct{}{}        // acquire
			defer func() { <-sem }() // release

			results[idx] = g.generatePage(pc)
		}(i, page)
	}

	wg.Wait()
	return results, nil
}

// GenerateToFiles renders pages and writes them to disk.
func (g *Generator) GenerateToFiles(pages []PageConfig) ([]Result, error) {
	results, err := g.Generate(pages)
	if err != nil {
		return nil, err
	}

	for i, r := range results {
		if r.Error != nil {
			continue
		}
		outPath := filepath.Join(g.config.OutputDir, r.OutFile)
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			results[i].Error = fmt.Errorf("failed to create directory for %s: %w", r.Path, err)
			continue
		}
		if err := os.WriteFile(outPath, []byte(r.HTML), 0644); err != nil {
			results[i].Error = fmt.Errorf("failed to write %s: %w", outPath, err)
		}
	}

	return results, nil
}

// generatePage renders a single page to HTML.
func (g *Generator) generatePage(pc PageConfig) Result {
	result := Result{
		Path:    pc.Path,
		OutFile: pathToFilename(pc.Path),
		Config:  pc,
	}

	opts := g.config.DefaultDocOpts
	if pc.DocumentOptions != nil {
		opts = *pc.DocumentOptions
	}

	html, err := g.renderer.RenderPageWithOptions(pc.Path, opts)
	if err != nil {
		if pc.Fallback == FallbackBlocking || pc.Fallback == FallbackStatic {
			result.HTML = g.generateFallbackHTML(pc, opts)
			result.Error = nil
		} else {
			result.Error = err
		}
		return result
	}

	result.HTML = html
	return result
}

// generateFallbackHTML creates a simple loading page for dynamic routes.
func (g *Generator) generateFallbackHTML(pc PageConfig, opts dom.DocumentOptions) string {
	var buf strings.Builder
	buf.WriteString("<!DOCTYPE html>\n<html>\n<head>\n")
	if opts.Title != "" {
		buf.WriteString(fmt.Sprintf("<title>%s</title>\n", opts.Title))
	}
	buf.WriteString("</head>\n<body>\n")
	buf.WriteString("<div id=\"app\"><p>Loading...</p></div>\n")
	buf.WriteString("</body>\n</html>")
	return buf.String()
}

// pathToFilename converts a URL path to a filename.
func pathToFilename(urlPath string) string {
	urlPath = strings.TrimPrefix(urlPath, "/")
	if urlPath == "" {
		return "index.html"
	}
	return filepath.Join(urlPath, "index.html")
}
