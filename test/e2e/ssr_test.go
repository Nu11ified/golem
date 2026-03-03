//go:build !js || !wasm

package e2e_test

import (
	"strings"
	"testing"

	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/internal/ssr"
	"github.com/Nu11ified/golem/router"
)

func TestSSRFullPageRender(t *testing.T) {
	r := router.NewRouter()
	r.AddSimpleRoute("/", func(params map[string]string) *dom.Element {
		return dom.H1(dom.Text("Welcome"))
	})

	opts := dom.DocumentOptions{
		Title: "Test App",
		Lang:  "en",
	}
	renderer := ssr.NewSSRRenderer(r, opts)

	html, err := renderer.RenderPage("/")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(html, "<!DOCTYPE html>") {
		t.Error("missing DOCTYPE")
	}
	if !strings.Contains(html, "<title>Test App</title>") {
		t.Error("missing title")
	}
	if !strings.Contains(html, "Welcome") {
		t.Error("missing page content")
	}
	if !strings.Contains(html, "wasm_exec.js") {
		t.Error("missing WASM bootstrap script")
	}
	if !strings.Contains(html, `<div id="app">`) {
		t.Error("missing app container")
	}
	if !strings.Contains(html, `lang="en"`) {
		t.Error("missing lang attribute")
	}
}

func TestSSRWithLayoutChain(t *testing.T) {
	r := router.NewRouter()
	r.AddSimpleRouteWithLayout("/page",
		func(params map[string]string) *dom.Element {
			return dom.NewElement("article", dom.Text("Page Content"))
		},
		func(child *dom.Element) *dom.Element {
			return dom.Div(
				dom.NewElement("header", dom.Text("Site Header")),
				child,
				dom.NewElement("footer", dom.Text("Site Footer")),
			)
		},
	)

	renderer := ssr.NewSSRRenderer(r, dom.DocumentOptions{Title: "Layout Test"})

	html, err := renderer.RenderPage("/page")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(html, "Site Header") {
		t.Error("missing layout header")
	}
	if !strings.Contains(html, "Page Content") {
		t.Error("missing page content")
	}
	if !strings.Contains(html, "Site Footer") {
		t.Error("missing layout footer")
	}

	headerIdx := strings.Index(html, "Site Header")
	contentIdx := strings.Index(html, "Page Content")
	footerIdx := strings.Index(html, "Site Footer")
	if headerIdx >= contentIdx || contentIdx >= footerIdx {
		t.Error("layout elements not in expected order")
	}
}

func TestSSRWithMetaTagsAndTitle(t *testing.T) {
	r := router.NewRouter()
	r.AddSimpleRoute("/", func(params map[string]string) *dom.Element {
		return dom.Div(dom.Text("Content"))
	})

	opts := dom.DocumentOptions{
		Title: "My App - Home",
		Lang:  "en",
		Meta: map[string]string{
			"description": "A test application",
			"viewport":    "width=device-width, initial-scale=1",
		},
	}

	renderer := ssr.NewSSRRenderer(r, opts)
	html, err := renderer.RenderPage("/")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(html, "<title>My App - Home</title>") {
		t.Error("missing title tag")
	}
	if !strings.Contains(html, `name="description"`) {
		t.Error("missing description meta tag")
	}
	if !strings.Contains(html, `content="A test application"`) {
		t.Error("missing description content")
	}
	if !strings.Contains(html, `name="viewport"`) {
		t.Error("missing viewport meta tag")
	}
}

func TestSSRWithWASMBootstrapScript(t *testing.T) {
	r := router.NewRouter()
	r.AddSimpleRoute("/", func(params map[string]string) *dom.Element {
		return dom.Div(dom.Text("App"))
	})

	renderer := ssr.NewSSRRenderer(r, dom.DocumentOptions{Title: "WASM App"})
	html, err := renderer.RenderPage("/")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(html, `<script src="/wasm_exec.js"></script>`) {
		t.Error("missing wasm_exec.js script tag")
	}
	if !strings.Contains(html, "WebAssembly.instantiateStreaming") {
		t.Error("missing WebAssembly instantiation code")
	}
	if !strings.Contains(html, "app.wasm") {
		t.Error("missing app.wasm reference")
	}
}

func TestSSRWithParallelSlotsRendered(t *testing.T) {
	r := router.NewRouter()
	r.AddRouteWithParallelSlots("/dashboard",
		func(params map[string]string) *dom.Element {
			return dom.Div(dom.Text("Main Content"))
		},
		map[string]func(params map[string]string) *dom.Element{
			"sidebar": func(params map[string]string) *dom.Element {
				return dom.NewElement("nav", dom.Text("Sidebar Nav"))
			},
		},
	)

	renderer := ssr.NewSSRRenderer(r, dom.DocumentOptions{Title: "Dashboard"})
	html, err := renderer.RenderPage("/dashboard")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(html, "Main Content") {
		t.Error("missing main content in SSR output")
	}
	if !strings.Contains(html, "Sidebar Nav") {
		t.Error("missing parallel slot content in SSR output")
	}
	if !strings.Contains(html, `data-golem-slot="sidebar"`) {
		t.Error("missing data-golem-slot attribute for parallel slot")
	}
}

func TestSSRErrorBoundaryRendering(t *testing.T) {
	r := router.NewRouter()
	r.AddRouteWithErrorBoundary("/error-page",
		func(params map[string]string) *dom.Element {
			panic("render failed")
		},
		func(err error) *dom.Element {
			return dom.Div(
				dom.H1(dom.Text("Something went wrong")),
				dom.P(dom.Text(err.Error())),
			)
		},
	)

	renderer := ssr.NewSSRRenderer(r, dom.DocumentOptions{Title: "Error Test"})
	html, err := renderer.RenderPage("/error-page")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(html, "Something went wrong") {
		t.Error("expected error boundary heading in SSR output")
	}
	if !strings.Contains(html, "render failed") {
		t.Error("expected error message in SSR output")
	}
	if !strings.Contains(html, "<title>Error Test</title>") {
		t.Error("expected title even in error case")
	}
}
