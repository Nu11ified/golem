package router_test

import (
	"testing"

	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/router"
)

// helper to create a dummy component that returns a distinguishable element
func makeComponent(name string) func(params map[string]string) *dom.Element {
	return func(params map[string]string) *dom.Element {
		return dom.Div(dom.Text(name))
	}
}

// TestRouter_InterceptingRoute_ClientNav verifies that client-side navigation
// to an intercepted path uses the intercepting component instead of the
// full-page component.
func TestRouter_InterceptingRoute_ClientNav(t *testing.T) {
	r := router.NewRouter()

	// Full page route for /photos/:id
	fullPage := &router.Route{
		Path:      "/photos/:id",
		Component: makeComponent("full-page"),
	}
	r.AddRoute(fullPage)

	// Intercepting route: when navigating client-side to /photos/:id,
	// render a modal overlay instead
	intercepting := &router.Route{
		Path:            "/photos/:id",
		Component:       makeComponent("modal-overlay"),
		IsIntercepting:  true,
		InterceptTarget: "/photos/:id",
	}
	r.AddRoute(intercepting)

	// Simulate client-side navigation via Push
	err := r.Push("/photos/42")
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	current := r.GetCurrentRoute()
	if current == nil {
		t.Fatal("expected current route to be set")
	}

	// Client-side nav should use the intercepting route
	if !current.IsIntercepting {
		t.Error("expected intercepting route to be used for client-side navigation")
	}
}

// TestRouter_InterceptingRoute_DirectNav verifies that direct URL access
// (e.g., page refresh, typing URL) uses the full page component, not the
// intercepting one.
func TestRouter_InterceptingRoute_DirectNav(t *testing.T) {
	r := router.NewRouter()

	// Full page route for /photos/:id
	fullPage := &router.Route{
		Path:      "/photos/:id",
		Component: makeComponent("full-page"),
	}
	r.AddRoute(fullPage)

	// Intercepting route
	intercepting := &router.Route{
		Path:            "/photos/:id",
		Component:       makeComponent("modal-overlay"),
		IsIntercepting:  true,
		InterceptTarget: "/photos/:id",
	}
	r.AddRoute(intercepting)

	// IsDirectNavigation is true by default (simulating fresh page load).
	// Call Navigate directly without Push to keep IsDirectNavigation=true.
	err := r.Navigate("/photos/42")
	if err != nil {
		t.Fatalf("Navigate failed: %v", err)
	}

	current := r.GetCurrentRoute()
	if current == nil {
		t.Fatal("expected current route to be set")
	}

	// Direct navigation should use the full page route, not the intercepting one
	if current.IsIntercepting {
		t.Error("expected full page route for direct navigation, got intercepting route")
	}
}

// TestRouter_InterceptingRoute_SameLevel tests the (.) convention which
// intercepts at the same directory level. For example, /photos/(.)photos/:id
// intercepts navigation to /photos/:id from within /photos.
func TestRouter_InterceptingRoute_SameLevel(t *testing.T) {
	r := router.NewRouter()

	// Full page route
	fullPage := &router.Route{
		Path:      "/photos/:id",
		Component: makeComponent("full-page"),
	}
	r.AddRoute(fullPage)

	// Same-level intercepting route: (.)photos/:id
	// The InterceptTarget is the resolved target path pattern
	intercepting := &router.Route{
		Path:            "/photos/:id",
		Component:       makeComponent("same-level-intercept"),
		IsIntercepting:  true,
		InterceptTarget: "/photos/:id",
	}
	r.AddRoute(intercepting)

	// Client-side navigation should resolve to intercepting route
	resolved := r.ResolveInterceptingRoute("/photos/1")
	if resolved == nil {
		t.Fatal("expected intercepting route to be resolved for same-level pattern")
	}
	if !resolved.IsIntercepting {
		t.Error("resolved route should be an intercepting route")
	}
}

// TestRouter_InterceptingRoute_OneUp tests the (..) convention which
// intercepts one level up. For example, /gallery/(..)photos/:id intercepts
// navigation to /photos/:id from within /gallery.
func TestRouter_InterceptingRoute_OneUp(t *testing.T) {
	r := router.NewRouter()

	// Full page route
	fullPage := &router.Route{
		Path:      "/photos/:id",
		Component: makeComponent("full-page"),
	}
	r.AddRoute(fullPage)

	// One-level-up intercepting route: (..)photos/:id resolves to /photos/:id
	intercepting := &router.Route{
		Path:            "/photos/:id",
		Component:       makeComponent("one-up-intercept"),
		IsIntercepting:  true,
		InterceptTarget: "/photos/:id",
	}
	r.AddRoute(intercepting)

	// Should resolve for a matching path
	resolved := r.ResolveInterceptingRoute("/photos/99")
	if resolved == nil {
		t.Fatal("expected intercepting route to be resolved for one-up pattern")
	}
	if !resolved.IsIntercepting {
		t.Error("resolved route should be an intercepting route")
	}
}

// TestRouter_InterceptingRoute_FromRoot tests the (...) convention which
// intercepts from the root level. For example, /a/b/c/(...)photos/:id
// intercepts navigation to /photos/:id from anywhere in the app.
func TestRouter_InterceptingRoute_FromRoot(t *testing.T) {
	r := router.NewRouter()

	// Full page route deep in the tree
	fullPage := &router.Route{
		Path:      "/admin/settings/photos/:id",
		Component: makeComponent("full-page"),
	}
	r.AddRoute(fullPage)

	// Root-level intercepting route: (...)admin/settings/photos/:id
	intercepting := &router.Route{
		Path:            "/admin/settings/photos/:id",
		Component:       makeComponent("root-intercept"),
		IsIntercepting:  true,
		InterceptTarget: "/admin/settings/photos/:id",
	}
	r.AddRoute(intercepting)

	// Should resolve the intercepting route
	resolved := r.ResolveInterceptingRoute("/admin/settings/photos/5")
	if resolved == nil {
		t.Fatal("expected intercepting route to be resolved for from-root pattern")
	}
	if !resolved.IsIntercepting {
		t.Error("resolved route should be an intercepting route")
	}
}

// TestRouter_InterceptingRoute_NoIntercept verifies that routes without
// intercept configuration use normal rendering -- ResolveInterceptingRoute
// returns nil and Navigate uses the standard route.
func TestRouter_InterceptingRoute_NoIntercept(t *testing.T) {
	r := router.NewRouter()

	// Regular route with no intercepting counterpart
	regular := &router.Route{
		Path:      "/about",
		Component: makeComponent("about-page"),
	}
	r.AddRoute(regular)

	// No intercepting route exists for /about
	resolved := r.ResolveInterceptingRoute("/about")
	if resolved != nil {
		t.Error("expected nil for path with no intercepting route")
	}

	// Client-side navigation should still work, using the regular route
	err := r.Push("/about")
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	current := r.GetCurrentRoute()
	if current == nil {
		t.Fatal("expected current route to be set")
	}
	if current.IsIntercepting {
		t.Error("expected regular route, not intercepting")
	}
	if current.Path != "/about" {
		t.Errorf("expected path /about, got %s", current.Path)
	}
}

// TestRouter_InterceptingRoute_ExactTargetMatch tests that
// ResolveInterceptingRoute correctly matches exact (non-parameterized)
// InterceptTarget strings.
func TestRouter_InterceptingRoute_ExactTargetMatch(t *testing.T) {
	r := router.NewRouter()

	regular := &router.Route{
		Path:      "/dashboard",
		Component: makeComponent("dashboard"),
	}
	r.AddRoute(regular)

	intercepting := &router.Route{
		Path:            "/dashboard",
		Component:       makeComponent("dashboard-modal"),
		IsIntercepting:  true,
		InterceptTarget: "/dashboard",
	}
	r.AddRoute(intercepting)

	resolved := r.ResolveInterceptingRoute("/dashboard")
	if resolved == nil {
		t.Fatal("expected intercepting route for exact target match")
	}
	if resolved.InterceptTarget != "/dashboard" {
		t.Errorf("expected InterceptTarget /dashboard, got %s", resolved.InterceptTarget)
	}
}

// TestRouter_InterceptingRoute_IsDirectNavigation_DefaultTrue verifies that
// a newly created router has IsDirectNavigation set to true.
func TestRouter_InterceptingRoute_IsDirectNavigation_DefaultTrue(t *testing.T) {
	r := router.NewRouter()
	if !r.IsDirectNavigation {
		t.Error("expected IsDirectNavigation to be true by default")
	}
}

// TestRouter_InterceptingRoute_PushSetsDirectNavFalse verifies that calling
// Push sets IsDirectNavigation to false.
func TestRouter_InterceptingRoute_PushSetsDirectNavFalse(t *testing.T) {
	r := router.NewRouter()
	r.AddRoute(&router.Route{
		Path:      "/test",
		Component: makeComponent("test"),
	})

	if !r.IsDirectNavigation {
		t.Fatal("expected IsDirectNavigation to be true before Push")
	}

	err := r.Push("/test")
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	if r.IsDirectNavigation {
		t.Error("expected IsDirectNavigation to be false after Push")
	}
}
