//go:build !js || !wasm

package e2e_test

import (
	"net/http"
	"testing"

	"github.com/Nu11ified/golem/internal/middleware"
)

func TestMiddlewarePipelineOrder(t *testing.T) {
	var order []string

	p := middleware.NewPipeline()

	p.Use(func(req *middleware.Request, next func(*middleware.Request) *middleware.Response) *middleware.Response {
		order = append(order, "m1-before")
		resp := next(req)
		order = append(order, "m1-after")
		return resp
	})

	p.Use(func(req *middleware.Request, next func(*middleware.Request) *middleware.Response) *middleware.Response {
		order = append(order, "m2-before")
		resp := next(req)
		order = append(order, "m2-after")
		return resp
	})

	p.Use(func(req *middleware.Request, next func(*middleware.Request) *middleware.Response) *middleware.Response {
		order = append(order, "m3-before")
		resp := next(req)
		order = append(order, "m3-after")
		return resp
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	mwReq := &middleware.Request{Request: req}

	p.Execute(mwReq, func(r *middleware.Request) *middleware.Response {
		order = append(order, "handler")
		return &middleware.Response{StatusCode: 200}
	})

	expected := []string{"m1-before", "m2-before", "m3-before", "handler", "m3-after", "m2-after", "m1-after"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d calls, got %d: %v", len(expected), len(order), order)
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("at index %d: expected %q, got %q", i, v, order[i])
		}
	}
}

func TestMiddlewareShortCircuit(t *testing.T) {
	var order []string

	p := middleware.NewPipeline()

	// Auth middleware that short-circuits
	p.Use(func(req *middleware.Request, next func(*middleware.Request) *middleware.Response) *middleware.Response {
		order = append(order, "auth-check")
		if req.Header.Get("Authorization") == "" {
			order = append(order, "auth-reject")
			return &middleware.Response{
				StatusCode: http.StatusUnauthorized,
				Headers:    map[string]string{"Content-Type": "text/plain"},
			}
		}
		order = append(order, "auth-pass")
		return next(req)
	})

	p.Use(func(req *middleware.Request, next func(*middleware.Request) *middleware.Response) *middleware.Response {
		order = append(order, "log-before")
		resp := next(req)
		order = append(order, "log-after")
		return resp
	})

	handler := func(r *middleware.Request) *middleware.Response {
		order = append(order, "handler")
		return &middleware.Response{StatusCode: 200}
	}

	// Request without auth - should be short-circuited
	req1, _ := http.NewRequest("GET", "/secret", nil)
	resp1 := p.Execute(&middleware.Request{Request: req1}, handler)

	if resp1.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp1.StatusCode)
	}

	expectedShort := []string{"auth-check", "auth-reject"}
	if len(order) != len(expectedShort) {
		t.Fatalf("expected %d calls for short-circuit, got %d: %v", len(expectedShort), len(order), order)
	}

	// Request with auth - should pass through
	order = nil
	req2, _ := http.NewRequest("GET", "/secret", nil)
	req2.Header.Set("Authorization", "Bearer token123")
	resp2 := p.Execute(&middleware.Request{Request: req2}, handler)

	if resp2.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp2.StatusCode)
	}

	expectedFull := []string{"auth-check", "auth-pass", "log-before", "handler", "log-after"}
	if len(order) != len(expectedFull) {
		t.Fatalf("expected %d calls for full pipeline, got %d: %v", len(expectedFull), len(order), order)
	}
	for i, v := range expectedFull {
		if order[i] != v {
			t.Errorf("at index %d: expected %q, got %q", i, v, order[i])
		}
	}
}

func TestMiddlewarePathMatching(t *testing.T) {
	var apiCalled, staticCalled bool

	p := middleware.NewPipeline()

	p.Use(func(req *middleware.Request, next func(*middleware.Request) *middleware.Response) *middleware.Response {
		apiCalled = true
		return next(req)
	}, middleware.Config{Matcher: []string{"/api/:path*"}})

	p.Use(func(req *middleware.Request, next func(*middleware.Request) *middleware.Response) *middleware.Response {
		staticCalled = true
		return next(req)
	}, middleware.Config{Matcher: []string{"/static/:path*"}})

	handler := func(r *middleware.Request) *middleware.Response {
		return &middleware.Response{StatusCode: 200}
	}

	// API path - only api middleware should fire
	req1, _ := http.NewRequest("GET", "/api/users", nil)
	p.Execute(&middleware.Request{Request: req1}, handler)

	if !apiCalled {
		t.Error("expected API middleware to be called for /api/users")
	}
	if staticCalled {
		t.Error("expected static middleware NOT to be called for /api/users")
	}
}
