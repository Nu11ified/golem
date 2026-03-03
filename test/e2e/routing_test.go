//go:build !js || !wasm

package e2e_test

import (
	"strings"
	"testing"

	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/router"
)

func setupTestRouter() *router.Router {
	r := router.NewRouter()

	r.AddSimpleRoute("/", func(params map[string]string) *dom.Element {
		return dom.Div(dom.Text("Home Page"))
	})

	r.AddSimpleRoute("/about", func(params map[string]string) *dom.Element {
		return dom.Div(dom.Text("About Page"))
	})

	r.AddSimpleRoute("/users/:id", func(params map[string]string) *dom.Element {
		return dom.Div(dom.Text("User " + params["id"]))
	})

	r.AddSimpleRoute("/users/:id/posts/:postId", func(params map[string]string) *dom.Element {
		return dom.Div(dom.Text("User " + params["id"] + " Post " + params["postId"]))
	})

	return r
}

func TestRouteRegistrationAndMatching(t *testing.T) {
	r := setupTestRouter()

	route, params := r.MatchRoute("/")
	if route == nil {
		t.Fatal("expected to match root route")
	}
	if len(params) != 0 {
		t.Errorf("expected no params for root, got %v", params)
	}

	el := route.Component(params)
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "Home Page") {
		t.Errorf("expected 'Home Page' in HTML, got: %s", html)
	}
}

func TestRouteMatchingAbout(t *testing.T) {
	r := setupTestRouter()

	route, params := r.MatchRoute("/about")
	if route == nil {
		t.Fatal("expected to match /about route")
	}
	if len(params) != 0 {
		t.Errorf("expected no params for /about, got %v", params)
	}

	el := route.Component(params)
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "About Page") {
		t.Errorf("expected 'About Page' in HTML, got: %s", html)
	}
}

func TestNestedLayoutWrapping(t *testing.T) {
	r := router.NewRouter()

	r.AddSimpleRouteWithLayout("/products",
		func(params map[string]string) *dom.Element {
			return dom.NewElement("main", dom.Text("Products"))
		},
		func(child *dom.Element) *dom.Element {
			return dom.Div(
				dom.NewElement("nav", dom.Text("Nav")),
				child,
			)
		},
	)

	route, params := r.MatchRoute("/products")
	if route == nil {
		t.Fatal("expected to match /products")
	}

	el, err := router.RenderWithErrorBoundary(route, params)
	if err != nil {
		t.Fatal(err)
	}

	wrapped := router.BuildLayoutChain(route, el)
	html := dom.RenderToHTML(wrapped)

	if !strings.Contains(html, "Nav") {
		t.Errorf("expected layout nav in HTML, got: %s", html)
	}
	if !strings.Contains(html, "Products") {
		t.Errorf("expected page content in HTML, got: %s", html)
	}
}

func TestDynamicRouteParameterExtraction(t *testing.T) {
	r := setupTestRouter()

	route, params := r.MatchRoute("/users/42")
	if route == nil {
		t.Fatal("expected to match /users/:id")
	}
	if params["id"] != "42" {
		t.Errorf("expected id=42, got id=%s", params["id"])
	}

	el := route.Component(params)
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "User 42") {
		t.Errorf("expected 'User 42' in HTML, got: %s", html)
	}
}

func TestDynamicRouteMultipleParams(t *testing.T) {
	r := setupTestRouter()

	route, params := r.MatchRoute("/users/7/posts/99")
	if route == nil {
		t.Fatal("expected to match /users/:id/posts/:postId")
	}
	if params["id"] != "7" {
		t.Errorf("expected id=7, got id=%s", params["id"])
	}
	if params["postId"] != "99" {
		t.Errorf("expected postId=99, got postId=%s", params["postId"])
	}

	el := route.Component(params)
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "User 7 Post 99") {
		t.Errorf("expected 'User 7 Post 99' in HTML, got: %s", html)
	}
}

func TestErrorBoundaryRendering(t *testing.T) {
	r := router.NewRouter()

	r.AddRouteWithErrorBoundary("/broken",
		func(params map[string]string) *dom.Element {
			panic("something went wrong")
		},
		func(err error) *dom.Element {
			return dom.Div(
				dom.H1(dom.Text("Error")),
				dom.P(dom.Text(err.Error())),
			)
		},
	)

	route, params := r.MatchRoute("/broken")
	if route == nil {
		t.Fatal("expected to match /broken")
	}

	el, err := router.RenderWithErrorBoundary(route, params)
	if err != nil {
		t.Fatal("expected error boundary to catch panic, got propagated error")
	}
	if el == nil {
		t.Fatal("expected error boundary element to be returned")
	}

	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "Error") {
		t.Errorf("expected 'Error' heading in HTML, got: %s", html)
	}
	if !strings.Contains(html, "something went wrong") {
		t.Errorf("expected error message in HTML, got: %s", html)
	}
}

func TestNotFoundHandler(t *testing.T) {
	r := router.NewRouter()
	r.AddSimpleRoute("/", func(params map[string]string) *dom.Element {
		return dom.Div(dom.Text("Home"))
	})
	r.NotFound(func() *dom.Element {
		return dom.Div(dom.H1(dom.Text("404 Not Found")))
	})

	route, _ := r.MatchRoute("/nonexistent")
	if route != nil {
		t.Fatal("expected no route for /nonexistent")
	}

	notFound := r.GetNotFoundHandler()
	if notFound == nil {
		t.Fatal("expected not-found handler to be set")
	}

	el := notFound()
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "404 Not Found") {
		t.Errorf("expected '404 Not Found' in HTML, got: %s", html)
	}
}

func TestRouteWithParallelSlots(t *testing.T) {
	r := router.NewRouter()

	r.AddRouteWithParallelSlots("/dashboard",
		func(params map[string]string) *dom.Element {
			return dom.Div(dom.Text("Dashboard Main"))
		},
		map[string]func(params map[string]string) *dom.Element{
			"sidebar": func(params map[string]string) *dom.Element {
				return dom.NewElement("aside", dom.Text("Sidebar"))
			},
			"header": func(params map[string]string) *dom.Element {
				return dom.NewElement("header", dom.Text("Header"))
			},
		},
	)

	route, params := r.MatchRoute("/dashboard")
	if route == nil {
		t.Fatal("expected to match /dashboard")
	}

	mainEl := route.Component(params)
	mainHTML := dom.RenderToHTML(mainEl)
	if !strings.Contains(mainHTML, "Dashboard Main") {
		t.Errorf("expected 'Dashboard Main' in HTML, got: %s", mainHTML)
	}

	slots := router.RenderParallelSlots(route, params)
	if len(slots) != 2 {
		t.Errorf("expected 2 parallel slots, got %d", len(slots))
	}

	sidebarEl, ok := slots["sidebar"]
	if !ok {
		t.Fatal("expected 'sidebar' slot")
	}
	if !strings.Contains(dom.RenderToHTML(sidebarEl), "Sidebar") {
		t.Error("expected 'Sidebar' in slot HTML")
	}

	headerEl, ok := slots["header"]
	if !ok {
		t.Fatal("expected 'header' slot")
	}
	if !strings.Contains(dom.RenderToHTML(headerEl), "Header") {
		t.Error("expected 'Header' in slot HTML")
	}
}
