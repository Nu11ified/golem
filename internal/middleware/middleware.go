package middleware

import (
	"net/http"
	"strings"
)

// Request wraps an HTTP request with convenience methods
type Request struct {
	*http.Request
	PathParams map[string]string
}

// Response wraps an HTTP response
type Response struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
	Redirect   string // if set, redirect to this URL
}

// Middleware processes a request and calls next to continue the chain
type Middleware func(req *Request, next func(*Request) *Response) *Response

// Config configures which paths a middleware applies to
type Config struct {
	Matcher []string // glob patterns like "/dashboard/:path*", "/api/:path*"
}

// Pipeline chains multiple middleware functions
type Pipeline struct {
	middlewares []middlewareEntry
}

type middlewareEntry struct {
	handler Middleware
	config  Config
}

// NewPipeline creates a new empty middleware pipeline.
func NewPipeline() *Pipeline {
	return &Pipeline{}
}

// Use adds a middleware to the pipeline with optional path matching.
// If no config is provided, the middleware applies to all paths.
func (p *Pipeline) Use(m Middleware, config ...Config) {
	entry := middlewareEntry{handler: m}
	if len(config) > 0 {
		entry.config = config[0]
	}
	p.middlewares = append(p.middlewares, entry)
}

// Execute runs the middleware pipeline for a given request.
// It filters middlewares by path matching, builds a chain of applicable
// middlewares, and executes the chain with the handler as the final function.
func (p *Pipeline) Execute(req *Request, handler func(*Request) *Response) *Response {
	// Filter middlewares that apply to this request path
	applicable := p.applicableMiddlewares(req.URL.Path)

	// Build chain from the inside out: handler is the innermost function,
	// and each middleware wraps the next one.
	current := handler
	for i := len(applicable) - 1; i >= 0; i-- {
		mw := applicable[i]
		// Capture mw and current in the closure
		next := current
		current = func(r *Request) *Response {
			return mw(r, next)
		}
	}

	return current(req)
}

// applicableMiddlewares returns the middleware handlers that match the given path.
func (p *Pipeline) applicableMiddlewares(path string) []Middleware {
	var result []Middleware
	for _, entry := range p.middlewares {
		if len(entry.config.Matcher) == 0 {
			// No matcher configured means the middleware applies to all paths
			result = append(result, entry.handler)
			continue
		}
		for _, pattern := range entry.config.Matcher {
			if MatchPath(path, pattern) {
				result = append(result, entry.handler)
				break
			}
		}
	}
	return result
}

// MatchPath checks if a request path matches a pattern.
//
// Supported pattern syntax:
//   - Exact match: "/about" matches only "/about"
//   - Wildcard: "/api/:path*" matches "/api", "/api/users", "/api/users/123"
//   - Single param: "/blog/:slug" matches "/blog/hello" but not "/blog/hello/comments"
//   - Mixed: "/users/:id/profile" matches "/users/42/profile"
func MatchPath(path string, pattern string) bool {
	pathParts := splitPath(path)
	patternParts := splitPath(pattern)

	pi := 0 // path index
	for i := 0; i < len(patternParts); i++ {
		part := patternParts[i]

		if strings.HasPrefix(part, ":") && strings.HasSuffix(part, "*") {
			// Wildcard parameter like :path* — matches zero or more remaining segments
			return true
		}

		if pi >= len(pathParts) {
			// Pattern has more segments than the path
			return false
		}

		if strings.HasPrefix(part, ":") {
			// Single segment parameter like :slug or :id — matches exactly one segment
			pi++
			continue
		}

		// Literal segment — must match exactly
		if pathParts[pi] != part {
			return false
		}
		pi++
	}

	// After consuming all pattern parts, the path should be fully consumed too
	return pi == len(pathParts)
}

// splitPath splits a URL path into non-empty segments.
// For example, "/api/users/123" becomes ["api", "users", "123"].
func splitPath(path string) []string {
	parts := strings.Split(path, "/")
	var result []string
	for _, p := range parts {
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// WithHeaders returns a middleware that adds the given headers to every response.
// Existing headers from the downstream handler are preserved; the provided
// headers are merged on top.
func WithHeaders(headers map[string]string) Middleware {
	return func(req *Request, next func(*Request) *Response) *Response {
		resp := next(req)
		if resp.Headers == nil {
			resp.Headers = make(map[string]string)
		}
		for k, v := range headers {
			resp.Headers[k] = v
		}
		return resp
	}
}

// WithAuth returns a middleware that checks authorization using the provided
// checker function. If the checker returns false, the middleware short-circuits
// with a 302 redirect to the given redirectTo URL.
func WithAuth(checker func(req *Request) bool, redirectTo string) Middleware {
	return func(req *Request, next func(*Request) *Response) *Response {
		if !checker(req) {
			return &Response{
				StatusCode: http.StatusFound,
				Redirect:   redirectTo,
				Headers:    map[string]string{"Location": redirectTo},
			}
		}
		return next(req)
	}
}
