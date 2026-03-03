//go:build !js || !wasm

// Package ssr provides server-side rendering for Golem applications.
// It renders pages to HTML on the server for initial page loads, enabling
// fast first-paint and SEO-friendly output. The rendered HTML includes
// hydration IDs and a WASM bootstrap script so the client can take over
// interactivity once the WebAssembly module loads.
package ssr

import (
	"fmt"
	"html"
	"sort"
	"strings"

	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/router"
)

// wasmBootstrapScript is the JavaScript snippet that loads the WASM module
// and hands off to the client-side application.
const wasmBootstrapScript = `<script src="/wasm_exec.js"></script>
<script>
if (!WebAssembly.instantiateStreaming) {
  WebAssembly.instantiateStreaming = async (resp, importObject) => {
    const source = await (await resp).arrayBuffer();
    return await WebAssembly.instantiate(source, importObject);
  };
}
const go = new Go();
WebAssembly.instantiateStreaming(fetch("/app.wasm"), go.importObject).then((result) => {
  go.run(result.instance);
});
</script>`

// SSRRenderer renders Golem pages to HTML on the server.
type SSRRenderer struct {
	router *router.Router
	opts   dom.DocumentOptions
}

// NewSSRRenderer creates a new SSR renderer with the given router and default
// document options.
func NewSSRRenderer(r *router.Router, opts dom.DocumentOptions) *SSRRenderer {
	return &SSRRenderer{
		router: r,
		opts:   opts,
	}
}

// RenderPage renders the page at the given path to a full HTML document string
// using the renderer's default DocumentOptions.
func (s *SSRRenderer) RenderPage(path string) (string, error) {
	return s.RenderPageWithOptions(path, s.opts)
}

// RenderPageWithOptions renders the page at the given path to a full HTML
// document string using the provided DocumentOptions.
func (s *SSRRenderer) RenderPageWithOptions(path string, opts dom.DocumentOptions) (string, error) {
	// Match the route
	route, params := s.router.MatchRoute(path)

	if route == nil {
		// Check for a custom not-found handler
		notFoundHandler := s.router.GetNotFoundHandler()
		if notFoundHandler != nil {
			element := notFoundHandler()
			return s.renderToDocument(element, nil, nil, opts), nil
		}
		return "", fmt.Errorf("route not found: %s", path)
	}

	// Render the component with error boundary support
	element, err := router.RenderWithErrorBoundary(route, params)
	if err != nil {
		return "", fmt.Errorf("render error: %w", err)
	}

	// Apply layout if present
	if route.Layout != nil && element != nil {
		element = router.BuildLayoutChain(route, element)
	}

	// Render parallel slots if present
	var parallelSlots map[string]*dom.Element
	if route.ParallelSlots != nil {
		parallelSlots = router.RenderParallelSlots(route, params)
	}

	return s.renderToDocument(element, parallelSlots, route, opts), nil
}

// renderToDocument assembles the full HTML document from the rendered content.
func (s *SSRRenderer) renderToDocument(
	content *dom.Element,
	parallelSlots map[string]*dom.Element,
	route *router.Route,
	opts dom.DocumentOptions,
) string {
	// Render the main content with hydration IDs
	var bodyHTML strings.Builder

	bodyHTML.WriteString(`<div id="app">`)

	if content != nil {
		bodyHTML.WriteString(dom.RenderToHTMLWithIDs(content))
	}

	// Render parallel slots in deterministic order
	if len(parallelSlots) > 0 {
		slotNames := make([]string, 0, len(parallelSlots))
		for name := range parallelSlots {
			slotNames = append(slotNames, name)
		}
		sort.Strings(slotNames)

		for _, name := range slotNames {
			el := parallelSlots[name]
			if el != nil {
				bodyHTML.WriteString(fmt.Sprintf(`<div data-golem-slot="%s">`, name))
				bodyHTML.WriteString(dom.RenderToHTMLWithIDs(el))
				bodyHTML.WriteString("</div>")
			}
		}
	}

	bodyHTML.WriteString("</div>")

	// Append WASM bootstrap script
	bodyHTML.WriteString("\n")
	bodyHTML.WriteString(wasmBootstrapScript)

	// Build the full HTML document
	return buildDocument(bodyHTML.String(), opts)
}

// buildDocument wraps body HTML content in a full HTML document.
// This is separate from dom.RenderDocument because SSR produces
// pre-rendered HTML strings (with hydration IDs) rather than Element trees.
func buildDocument(bodyContent string, opts dom.DocumentOptions) string {
	lang := opts.Lang
	if lang == "" {
		lang = "en"
	}

	var buf strings.Builder
	buf.WriteString("<!DOCTYPE html>\n")
	buf.WriteString(fmt.Sprintf("<html lang=\"%s\">\n", html.EscapeString(lang)))
	buf.WriteString("<head>\n")
	buf.WriteString("<meta charset=\"utf-8\" />\n")

	if opts.Title != "" {
		buf.WriteString(fmt.Sprintf("<title>%s</title>\n", html.EscapeString(opts.Title)))
	}

	if len(opts.Meta) > 0 {
		metaKeys := make([]string, 0, len(opts.Meta))
		for k := range opts.Meta {
			metaKeys = append(metaKeys, k)
		}
		sort.Strings(metaKeys)
		for _, name := range metaKeys {
			content := opts.Meta[name]
			buf.WriteString(fmt.Sprintf("<meta name=\"%s\" content=\"%s\" />\n",
				html.EscapeString(name), html.EscapeString(content)))
		}
	}

	for _, href := range opts.Styles {
		buf.WriteString(fmt.Sprintf("<link rel=\"stylesheet\" href=\"%s\" />\n", html.EscapeString(href)))
	}

	if opts.WasmExecPath != "" {
		buf.WriteString(fmt.Sprintf("<script src=\"%s\"></script>\n", html.EscapeString(opts.WasmExecPath)))
	}

	for _, src := range opts.Scripts {
		buf.WriteString(fmt.Sprintf("<script src=\"%s\"></script>\n", html.EscapeString(src)))
	}

	buf.WriteString("</head>\n")
	buf.WriteString("<body>\n")
	buf.WriteString(bodyContent)
	buf.WriteByte('\n')
	buf.WriteString("</body>\n")
	buf.WriteString("</html>")

	return buf.String()
}
