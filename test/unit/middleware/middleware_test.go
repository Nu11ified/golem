package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Nu11ified/golem/internal/middleware"
)

func makeRequest(method, path string) *middleware.Request {
	req := httptest.NewRequest(method, path, nil)
	return &middleware.Request{
		Request:    req,
		PathParams: make(map[string]string),
	}
}

func simpleHandler(req *middleware.Request) *middleware.Response {
	return &middleware.Response{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "text/plain"},
		Body:       []byte("handler response"),
	}
}

// TestPipeline_NoMiddleware verifies that a request passes straight through
// to the handler when no middleware is registered.
func TestPipeline_NoMiddleware(t *testing.T) {
	p := middleware.NewPipeline()
	req := makeRequest("GET", "/hello")

	resp := p.Execute(req, simpleHandler)

	if resp == nil {
		t.Fatal("expected a response, got nil")
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if string(resp.Body) != "handler response" {
		t.Errorf("expected 'handler response', got %q", string(resp.Body))
	}
}

// TestPipeline_SingleMiddleware verifies that a single middleware can modify
// the request before it reaches the handler.
func TestPipeline_SingleMiddleware(t *testing.T) {
	p := middleware.NewPipeline()

	// Middleware that adds a custom header to the request
	p.Use(func(req *middleware.Request, next func(*middleware.Request) *middleware.Response) *middleware.Response {
		req.Header.Set("X-Modified", "true")
		return next(req)
	})

	var receivedHeader string
	handler := func(req *middleware.Request) *middleware.Response {
		receivedHeader = req.Header.Get("X-Modified")
		return &middleware.Response{
			StatusCode: http.StatusOK,
			Body:       []byte("ok"),
		}
	}

	req := makeRequest("GET", "/test")
	p.Execute(req, handler)

	if receivedHeader != "true" {
		t.Errorf("expected handler to receive X-Modified header 'true', got %q", receivedHeader)
	}
}

// TestPipeline_ChainedMiddleware verifies that two middlewares execute in
// the correct order (first registered runs first).
func TestPipeline_ChainedMiddleware(t *testing.T) {
	p := middleware.NewPipeline()

	var order []string

	p.Use(func(req *middleware.Request, next func(*middleware.Request) *middleware.Response) *middleware.Response {
		order = append(order, "first")
		resp := next(req)
		order = append(order, "first-after")
		return resp
	})

	p.Use(func(req *middleware.Request, next func(*middleware.Request) *middleware.Response) *middleware.Response {
		order = append(order, "second")
		resp := next(req)
		order = append(order, "second-after")
		return resp
	})

	handler := func(req *middleware.Request) *middleware.Response {
		order = append(order, "handler")
		return &middleware.Response{StatusCode: http.StatusOK}
	}

	req := makeRequest("GET", "/test")
	p.Execute(req, handler)

	expected := []string{"first", "second", "handler", "second-after", "first-after"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d execution steps, got %d: %v", len(expected), len(order), order)
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("step %d: expected %q, got %q", i, v, order[i])
		}
	}
}

// TestPipeline_MiddlewareShortCircuit verifies that a middleware can return
// a response without calling next, short-circuiting the pipeline (e.g., auth failure).
func TestPipeline_MiddlewareShortCircuit(t *testing.T) {
	p := middleware.NewPipeline()

	handlerCalled := false

	// Auth middleware that rejects the request
	p.Use(func(req *middleware.Request, next func(*middleware.Request) *middleware.Response) *middleware.Response {
		return &middleware.Response{
			StatusCode: http.StatusUnauthorized,
			Body:       []byte("unauthorized"),
		}
	})

	handler := func(req *middleware.Request) *middleware.Response {
		handlerCalled = true
		return &middleware.Response{StatusCode: http.StatusOK}
	}

	req := makeRequest("GET", "/protected")
	resp := p.Execute(req, handler)

	if handlerCalled {
		t.Error("expected handler NOT to be called when middleware short-circuits")
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
	}
	if string(resp.Body) != "unauthorized" {
		t.Errorf("expected 'unauthorized' body, got %q", string(resp.Body))
	}
}

// TestPipeline_PathMatching verifies that a middleware with a path matcher
// is only applied when the request path matches.
func TestPipeline_PathMatching(t *testing.T) {
	p := middleware.NewPipeline()

	middlewareCalled := false

	p.Use(func(req *middleware.Request, next func(*middleware.Request) *middleware.Response) *middleware.Response {
		middlewareCalled = true
		req.Header.Set("X-API", "true")
		return next(req)
	}, middleware.Config{
		Matcher: []string{"/api/:path*"},
	})

	handler := func(req *middleware.Request) *middleware.Response {
		return &middleware.Response{
			StatusCode: http.StatusOK,
			Body:       []byte("ok"),
		}
	}

	// Request to /api/users should trigger the middleware
	req := makeRequest("GET", "/api/users")
	p.Execute(req, handler)

	if !middlewareCalled {
		t.Error("expected middleware to be called for /api/users")
	}
}

// TestPipeline_PathNotMatching verifies that a middleware with a specific
// matcher does NOT apply to paths that don't match.
func TestPipeline_PathNotMatching(t *testing.T) {
	p := middleware.NewPipeline()

	middlewareCalled := false

	p.Use(func(req *middleware.Request, next func(*middleware.Request) *middleware.Response) *middleware.Response {
		middlewareCalled = true
		return next(req)
	}, middleware.Config{
		Matcher: []string{"/api/:path*"},
	})

	handler := func(req *middleware.Request) *middleware.Response {
		return &middleware.Response{
			StatusCode: http.StatusOK,
			Body:       []byte("ok"),
		}
	}

	// Request to /about should NOT trigger the middleware
	req := makeRequest("GET", "/about")
	p.Execute(req, handler)

	if middlewareCalled {
		t.Error("expected middleware NOT to be called for /about")
	}
}

// TestMatchPath_Exact verifies exact path matching.
func TestMatchPath_Exact(t *testing.T) {
	tests := []struct {
		path    string
		pattern string
		match   bool
	}{
		{"/about", "/about", true},
		{"/about", "/contact", false},
		{"/", "/", true},
		{"/about/us", "/about", false},
	}

	for _, tt := range tests {
		result := middleware.MatchPath(tt.path, tt.pattern)
		if result != tt.match {
			t.Errorf("MatchPath(%q, %q) = %v, want %v", tt.path, tt.pattern, result, tt.match)
		}
	}
}

// TestMatchPath_Wildcard verifies wildcard path matching with :path*.
func TestMatchPath_Wildcard(t *testing.T) {
	tests := []struct {
		path    string
		pattern string
		match   bool
	}{
		{"/api/users", "/api/:path*", true},
		{"/api/posts/1", "/api/:path*", true},
		{"/api", "/api/:path*", true},
		{"/api/", "/api/:path*", true},
		{"/dashboard", "/dashboard/:path*", true},
		{"/dashboard/settings", "/dashboard/:path*", true},
		{"/dashboard/users/123", "/dashboard/:path*", true},
		{"/other/route", "/api/:path*", false},
		{"/apifoo", "/api/:path*", false},
	}

	for _, tt := range tests {
		result := middleware.MatchPath(tt.path, tt.pattern)
		if result != tt.match {
			t.Errorf("MatchPath(%q, %q) = %v, want %v", tt.path, tt.pattern, result, tt.match)
		}
	}
}

// TestMatchPath_Param verifies single-segment parameter matching with :param.
func TestMatchPath_Param(t *testing.T) {
	tests := []struct {
		path    string
		pattern string
		match   bool
	}{
		{"/blog/hello", "/blog/:slug", true},
		{"/blog/my-post", "/blog/:slug", true},
		{"/blog/hello/comments", "/blog/:slug", false},
		{"/blog/", "/blog/:slug", false},
		{"/blog", "/blog/:slug", false},
		{"/users/42/profile", "/users/:id/profile", true},
		{"/users/42/settings", "/users/:id/profile", false},
	}

	for _, tt := range tests {
		result := middleware.MatchPath(tt.path, tt.pattern)
		if result != tt.match {
			t.Errorf("MatchPath(%q, %q) = %v, want %v", tt.path, tt.pattern, result, tt.match)
		}
	}
}

// TestPipeline_RedirectResponse verifies that middleware can set a Redirect
// field for a 302 redirect.
func TestPipeline_RedirectResponse(t *testing.T) {
	p := middleware.NewPipeline()

	// Middleware that redirects unauthenticated users
	p.Use(func(req *middleware.Request, next func(*middleware.Request) *middleware.Response) *middleware.Response {
		if req.Header.Get("Authorization") == "" {
			return &middleware.Response{
				StatusCode: http.StatusFound,
				Redirect:   "/login",
				Headers:    map[string]string{"Location": "/login"},
			}
		}
		return next(req)
	})

	handler := func(req *middleware.Request) *middleware.Response {
		return &middleware.Response{
			StatusCode: http.StatusOK,
			Body:       []byte("protected content"),
		}
	}

	// Request without auth should be redirected
	req := makeRequest("GET", "/dashboard")
	resp := p.Execute(req, handler)

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected status 302, got %d", resp.StatusCode)
	}
	if resp.Redirect != "/login" {
		t.Errorf("expected redirect to '/login', got %q", resp.Redirect)
	}
	if resp.Headers["Location"] != "/login" {
		t.Errorf("expected Location header '/login', got %q", resp.Headers["Location"])
	}
}

// TestWithHeaders verifies the WithHeaders helper middleware.
func TestWithHeaders(t *testing.T) {
	p := middleware.NewPipeline()

	p.Use(middleware.WithHeaders(map[string]string{
		"X-Frame-Options":  "DENY",
		"X-Custom-Header":  "golem",
	}))

	handler := func(req *middleware.Request) *middleware.Response {
		return &middleware.Response{
			StatusCode: http.StatusOK,
			Headers:    map[string]string{"Content-Type": "text/html"},
			Body:       []byte("page"),
		}
	}

	req := makeRequest("GET", "/")
	resp := p.Execute(req, handler)

	if resp.Headers["X-Frame-Options"] != "DENY" {
		t.Errorf("expected X-Frame-Options 'DENY', got %q", resp.Headers["X-Frame-Options"])
	}
	if resp.Headers["X-Custom-Header"] != "golem" {
		t.Errorf("expected X-Custom-Header 'golem', got %q", resp.Headers["X-Custom-Header"])
	}
	// Original headers should be preserved
	if resp.Headers["Content-Type"] != "text/html" {
		t.Errorf("expected Content-Type 'text/html', got %q", resp.Headers["Content-Type"])
	}
}

// TestWithAuth_Authorized verifies that WithAuth allows authorized requests through.
func TestWithAuth_Authorized(t *testing.T) {
	p := middleware.NewPipeline()

	checker := func(req *middleware.Request) bool {
		return req.Header.Get("Authorization") == "Bearer valid-token"
	}

	p.Use(middleware.WithAuth(checker, "/login"))

	handlerCalled := false
	handler := func(req *middleware.Request) *middleware.Response {
		handlerCalled = true
		return &middleware.Response{
			StatusCode: http.StatusOK,
			Body:       []byte("secret data"),
		}
	}

	req := makeRequest("GET", "/protected")
	req.Header.Set("Authorization", "Bearer valid-token")
	resp := p.Execute(req, handler)

	if !handlerCalled {
		t.Error("expected handler to be called for authorized request")
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

// TestWithAuth_Unauthorized verifies that WithAuth redirects unauthorized requests.
func TestWithAuth_Unauthorized(t *testing.T) {
	p := middleware.NewPipeline()

	checker := func(req *middleware.Request) bool {
		return req.Header.Get("Authorization") == "Bearer valid-token"
	}

	p.Use(middleware.WithAuth(checker, "/login"))

	handlerCalled := false
	handler := func(req *middleware.Request) *middleware.Response {
		handlerCalled = true
		return &middleware.Response{StatusCode: http.StatusOK}
	}

	req := makeRequest("GET", "/protected")
	// No Authorization header set
	resp := p.Execute(req, handler)

	if handlerCalled {
		t.Error("expected handler NOT to be called for unauthorized request")
	}
	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected status 302, got %d", resp.StatusCode)
	}
	if resp.Redirect != "/login" {
		t.Errorf("expected redirect to '/login', got %q", resp.Redirect)
	}
}

// TestPipeline_MultipleMatcherPatterns verifies that a middleware with multiple
// matcher patterns matches any of them.
func TestPipeline_MultipleMatcherPatterns(t *testing.T) {
	callCount := 0

	p := middleware.NewPipeline()
	p.Use(func(req *middleware.Request, next func(*middleware.Request) *middleware.Response) *middleware.Response {
		callCount++
		return next(req)
	}, middleware.Config{
		Matcher: []string{"/api/:path*", "/dashboard/:path*"},
	})

	handler := func(req *middleware.Request) *middleware.Response {
		return &middleware.Response{StatusCode: http.StatusOK}
	}

	// Should match /api/users
	p.Execute(makeRequest("GET", "/api/users"), handler)
	if callCount != 1 {
		t.Errorf("expected middleware called 1 time, got %d", callCount)
	}

	// Should match /dashboard/settings
	p.Execute(makeRequest("GET", "/dashboard/settings"), handler)
	if callCount != 2 {
		t.Errorf("expected middleware called 2 times, got %d", callCount)
	}

	// Should NOT match /blog/post
	p.Execute(makeRequest("GET", "/blog/post"), handler)
	if callCount != 2 {
		t.Errorf("expected middleware still called 2 times, got %d", callCount)
	}
}

// TestPipeline_MiddlewareModifiesResponse verifies that middleware can modify
// the response returned by downstream handlers.
func TestPipeline_MiddlewareModifiesResponse(t *testing.T) {
	p := middleware.NewPipeline()

	// Middleware that adds a response header after the handler runs
	p.Use(func(req *middleware.Request, next func(*middleware.Request) *middleware.Response) *middleware.Response {
		resp := next(req)
		if resp.Headers == nil {
			resp.Headers = make(map[string]string)
		}
		resp.Headers["X-Powered-By"] = "Golem"
		return resp
	})

	handler := func(req *middleware.Request) *middleware.Response {
		return &middleware.Response{
			StatusCode: http.StatusOK,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       []byte(`{"ok":true}`),
		}
	}

	req := makeRequest("GET", "/")
	resp := p.Execute(req, handler)

	if resp.Headers["X-Powered-By"] != "Golem" {
		t.Errorf("expected X-Powered-By 'Golem', got %q", resp.Headers["X-Powered-By"])
	}
	if resp.Headers["Content-Type"] != "application/json" {
		t.Errorf("expected Content-Type preserved, got %q", resp.Headers["Content-Type"])
	}
}
