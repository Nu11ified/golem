//go:build !js || !wasm

package router_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/router"
)

func TestRouter_LayoutWrapping(t *testing.T) {
	r := router.NewRouter()

	layout := func(children *dom.Element) *dom.Element {
		return dom.Div(dom.Class("layout"), children)
	}

	page := func(params map[string]string) *dom.Element {
		return dom.P("page content")
	}

	route := &router.Route{
		Path:      "/",
		Component: page,
		Layout:    layout,
	}

	r.AddRoute(route)

	// Test that BuildLayoutChain wraps correctly
	content := page(nil)
	wrapped := r.BuildLayoutChain(route, content)

	html := dom.RenderToHTML(wrapped)
	if html == "" {
		t.Error("expected non-empty HTML")
	}
	// The layout div should wrap the page content
	// We verify structure through the element tree
	if wrapped.Type != "div" {
		t.Errorf("expected div wrapper from layout, got %s", wrapped.Type)
	}
}

func TestRouter_NestedLayouts(t *testing.T) {
	r := router.NewRouter()

	rootLayout := func(children *dom.Element) *dom.Element {
		return dom.Div(dom.Class("root-layout"), children)
	}

	blogLayout := func(children *dom.Element) *dom.Element {
		return dom.Div(dom.Class("blog-layout"), children)
	}

	rootRoute := &router.Route{
		Path:   "/",
		Layout: rootLayout,
	}

	blogRoute := &router.Route{
		Path:        "/blog",
		Layout:      blogLayout,
		ParentRoute: rootRoute,
		Component: func(params map[string]string) *dom.Element {
			return dom.P("blog page")
		},
	}

	r.AddRoute(rootRoute)
	r.AddRoute(blogRoute)

	content := blogRoute.Component(nil)
	wrapped := r.BuildLayoutChain(blogRoute, content)

	// Should be: rootLayout(blogLayout(content))
	// outer div should be root-layout
	if wrapped.Type != "div" {
		t.Errorf("expected div, got %s", wrapped.Type)
	}

	html := dom.RenderToHTML(wrapped)
	if !strings.Contains(html, "root-layout") {
		t.Errorf("expected root-layout class in HTML, got %s", html)
	}
	if !strings.Contains(html, "blog-layout") {
		t.Errorf("expected blog-layout class in HTML, got %s", html)
	}
	if !strings.Contains(html, "blog page") {
		t.Errorf("expected blog page content in HTML, got %s", html)
	}
}

func TestRouter_ErrorBoundary(t *testing.T) {
	r := router.NewRouter()

	errorHandler := func(err error) *dom.Element {
		return dom.Div(dom.Class("error"), dom.P(err.Error()))
	}

	panickyPage := func(params map[string]string) *dom.Element {
		panic(errors.New("something went wrong"))
	}

	route := &router.Route{
		Path:         "/broken",
		Component:    panickyPage,
		ErrorHandler: errorHandler,
	}

	r.AddRoute(route)

	// RenderWithErrorBoundary should catch the panic
	result := r.RenderWithErrorBoundary(route, nil)
	if result == nil {
		t.Fatal("expected error fallback element, got nil")
	}

	html := dom.RenderToHTML(result)
	if html == "" {
		t.Error("expected non-empty error HTML")
	}
	if !strings.Contains(html, "something went wrong") {
		t.Errorf("expected error message in HTML, got %s", html)
	}
}

func TestRouter_ErrorBoundaryNoError(t *testing.T) {
	r := router.NewRouter()

	errorHandler := func(err error) *dom.Element {
		return dom.Div(dom.Class("error"), dom.P(err.Error()))
	}

	normalPage := func(params map[string]string) *dom.Element {
		return dom.P("works fine")
	}

	route := &router.Route{
		Path:         "/good",
		Component:    normalPage,
		ErrorHandler: errorHandler,
	}

	result := r.RenderWithErrorBoundary(route, nil)
	if result == nil {
		t.Fatal("expected page element, got nil")
	}
	if result.Type != "p" {
		t.Errorf("expected p element, got %s", result.Type)
	}
}

func TestRouter_ErrorBoundaryStringPanic(t *testing.T) {
	r := router.NewRouter()

	errorHandler := func(err error) *dom.Element {
		return dom.Div(dom.Class("error"), dom.P(err.Error()))
	}

	panickyPage := func(params map[string]string) *dom.Element {
		panic("string panic value")
	}

	route := &router.Route{
		Path:         "/broken-string",
		Component:    panickyPage,
		ErrorHandler: errorHandler,
	}

	r.AddRoute(route)

	result := r.RenderWithErrorBoundary(route, nil)
	if result == nil {
		t.Fatal("expected error fallback element, got nil")
	}

	html := dom.RenderToHTML(result)
	if !strings.Contains(html, "string panic value") {
		t.Errorf("expected string panic message in HTML, got %s", html)
	}
}

func TestRouter_TemplateWrapping(t *testing.T) {
	r := router.NewRouter()

	template := func(children *dom.Element) *dom.Element {
		return dom.Div(dom.Class("template"), children)
	}

	page := func(params map[string]string) *dom.Element {
		return dom.P("template page")
	}

	route := &router.Route{
		Path:      "/templated",
		Component: page,
		Template:  template,
	}

	r.AddRoute(route)

	// Manually apply template like the router does
	content := page(nil)
	if route.Template != nil {
		content = route.Template(content)
	}

	html := dom.RenderToHTML(content)
	if !strings.Contains(html, "template") {
		t.Errorf("expected template class in HTML, got %s", html)
	}
	if !strings.Contains(html, "template page") {
		t.Errorf("expected page content in HTML, got %s", html)
	}
}

func TestRouter_AddRouteWithLayout(t *testing.T) {
	r := router.NewRouter()

	rootLayout := func(children *dom.Element) *dom.Element {
		return dom.Div(dom.Class("root"), children)
	}

	parentRoute := &router.Route{
		Path:   "/",
		Layout: rootLayout,
	}

	childRoute := &router.Route{
		Path: "/child",
		Component: func(params map[string]string) *dom.Element {
			return dom.P("child page")
		},
	}

	r.AddRoute(parentRoute)
	r.AddRouteWithLayout(childRoute, parentRoute)

	if childRoute.ParentRoute != parentRoute {
		t.Error("expected child route to have parent set")
	}

	content := childRoute.Component(nil)
	wrapped := r.BuildLayoutChain(childRoute, content)

	html := dom.RenderToHTML(wrapped)
	if !strings.Contains(html, "root") {
		t.Errorf("expected root layout class in HTML, got %s", html)
	}
	if !strings.Contains(html, "child page") {
		t.Errorf("expected child page content in HTML, got %s", html)
	}
}

func TestRouter_BuildLayoutChainNilRoute(t *testing.T) {
	r := router.NewRouter()

	content := dom.P("test")
	result := r.BuildLayoutChain(nil, content)

	if result != content {
		t.Error("expected content returned unchanged for nil route")
	}
}

func TestRouter_BuildLayoutChainNilContent(t *testing.T) {
	r := router.NewRouter()

	route := &router.Route{
		Path: "/",
		Layout: func(children *dom.Element) *dom.Element {
			return dom.Div(children)
		},
	}

	result := r.BuildLayoutChain(route, nil)
	if result != nil {
		t.Error("expected nil returned for nil content")
	}
}

func TestRouter_BuildLayoutChainNoLayout(t *testing.T) {
	r := router.NewRouter()

	route := &router.Route{
		Path: "/",
	}

	content := dom.P("no layout")
	result := r.BuildLayoutChain(route, content)

	if result != content {
		t.Error("expected content returned unchanged when route has no layout")
	}
}

func TestRouter_ParallelSlotsField(t *testing.T) {
	route := &router.Route{
		Path: "/dashboard",
		ParallelSlots: map[string]*router.Route{
			"sidebar": {
				Path: "/dashboard/@sidebar",
				Component: func(params map[string]string) *dom.Element {
					return dom.Div(dom.Class("sidebar"), dom.P("sidebar content"))
				},
			},
			"main": {
				Path: "/dashboard/@main",
				Component: func(params map[string]string) *dom.Element {
					return dom.Div(dom.Class("main"), dom.P("main content"))
				},
			},
		},
	}

	if len(route.ParallelSlots) != 2 {
		t.Errorf("expected 2 parallel slots, got %d", len(route.ParallelSlots))
	}

	sidebarSlot := route.ParallelSlots["sidebar"]
	if sidebarSlot == nil {
		t.Fatal("expected sidebar slot to exist")
	}

	sidebarContent := sidebarSlot.Component(nil)
	html := dom.RenderToHTML(sidebarContent)
	if !strings.Contains(html, "sidebar content") {
		t.Errorf("expected sidebar content in HTML, got %s", html)
	}
}

func TestRouter_LayoutWithErrorBoundaryCombined(t *testing.T) {
	r := router.NewRouter()

	layout := func(children *dom.Element) *dom.Element {
		return dom.Div(dom.Class("layout"), children)
	}

	errorHandler := func(err error) *dom.Element {
		return dom.Div(dom.Class("error"), dom.P(err.Error()))
	}

	panickyPage := func(params map[string]string) *dom.Element {
		panic(errors.New("layout error test"))
	}

	route := &router.Route{
		Path:         "/combined",
		Component:    panickyPage,
		Layout:       layout,
		ErrorHandler: errorHandler,
	}

	r.AddRoute(route)

	// First render with error boundary
	result := r.RenderWithErrorBoundary(route, nil)
	if result == nil {
		t.Fatal("expected error fallback element, got nil")
	}

	// Then apply layout chain
	wrapped := r.BuildLayoutChain(route, result)
	html := dom.RenderToHTML(wrapped)

	if !strings.Contains(html, "layout") {
		t.Errorf("expected layout class in HTML, got %s", html)
	}
	if !strings.Contains(html, "layout error test") {
		t.Errorf("expected error message in HTML, got %s", html)
	}
}
