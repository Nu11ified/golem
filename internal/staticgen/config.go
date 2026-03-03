//go:build !js || !wasm

package staticgen

import (
	"time"

	"github.com/Nu11ified/golem/dom"
)

// FallbackMode defines behavior for dynamic routes during static generation.
type FallbackMode int

const (
	// FallbackNone means no fallback — errors are propagated.
	FallbackNone FallbackMode = iota
	// FallbackBlocking means render on first request, then cache.
	FallbackBlocking
	// FallbackStatic means generate a static loading page.
	FallbackStatic
)

// PageConfig configures static generation for a single page.
type PageConfig struct {
	Path            string
	Revalidate      time.Duration        // 0 means no revalidation (purely static)
	Fallback        FallbackMode
	DocumentOptions *dom.DocumentOptions // per-page overrides, nil uses defaults
}

// GeneratorConfig configures the static page generator.
type GeneratorConfig struct {
	OutputDir      string
	Concurrency    int
	DefaultDocOpts dom.DocumentOptions
}
