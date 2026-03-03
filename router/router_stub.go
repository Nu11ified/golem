//go:build !js || !wasm

package router

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Nu11ified/golem/dom"
)

// Route represents a single route
type Route struct {
	Path          string
	Component     func(params map[string]string) *dom.Element
	Guards        []Guard
	Children      []*Route
	Meta          map[string]interface{}
	Name          string
	Redirect      string
	Regex         *regexp.Regexp
	ParamNames    []string
	Layout        func(content *dom.Element) *dom.Element
	ErrorHandler  func(err error) *dom.Element
	ParallelSlots map[string]func(params map[string]string) *dom.Element
}

// Guard represents a route guard
type Guard func(to *Route, from *Route, params map[string]string) bool

// Router manages routing
type Router struct {
	routes          []*Route
	currentRoute    *Route
	currentParams   map[string]string
	beforeEach      []Guard
	afterEach       []func(*Route, *Route)
	notFoundHandler func() *dom.Element
	errorHandler    func(error) *dom.Element
	baseURL         string
	mode            RouterMode
	container       string
}

// RouterMode defines routing modes
type RouterMode int

const (
	HashMode RouterMode = iota
	HistoryMode
)

// LinkComponent for navigation
type LinkComponent struct {
	To     string
	Class  string
	Text   string
	Router *Router
}

// TransitionHook is a route transition hook
type TransitionHook func(to *Route, from *Route, next func())

// Transition manages route transitions
type Transition struct {
	hooks []TransitionHook
}

// Guards provides common route guards
type Guards struct{}

// NewRouter creates a new router instance
func NewRouter() *Router {
	return &Router{
		routes:        make([]*Route, 0),
		currentParams: make(map[string]string),
		beforeEach:    make([]Guard, 0),
		afterEach:     make([]func(*Route, *Route), 0),
		mode:          HashMode,
		container:     "#router-outlet",
	}
}

func (r *Router) SetMode(mode RouterMode) *Router      { return r }
func (r *Router) SetContainer(selector string) *Router  { return r }
func (r *Router) SetBaseURL(baseURL string) *Router     { return r }

// AddRoute adds a route to the router with route compilation
func (r *Router) AddRoute(route *Route) *Router {
	r.compileRoute(route)
	r.routes = append(r.routes, route)
	return r
}

// AddSimpleRoute creates and adds a new route
func (r *Router) AddSimpleRoute(path string, component func(params map[string]string) *dom.Element) *Router {
	return r.AddRoute(&Route{
		Path:      path,
		Component: component,
	})
}

// RouteWithName creates and adds a named route
func (r *Router) RouteWithName(name, path string, component func(params map[string]string) *dom.Element) *Router {
	return r.AddRoute(&Route{
		Name:      name,
		Path:      path,
		Component: component,
	})
}

// RouteGroup creates a route group with shared guards
func (r *Router) RouteGroup(prefix string, guards []Guard, routes []*Route) *Router {
	for _, route := range routes {
		route.Path = prefix + route.Path
		route.Guards = append(guards, route.Guards...)
		r.AddRoute(route)
	}
	return r
}

func (r *Router) BeforeEach(guard Guard) *Router                   { return r }
func (r *Router) AfterEach(hook func(*Route, *Route)) *Router     { return r }

// NotFound sets the 404 handler
func (r *Router) NotFound(handler func() *dom.Element) *Router {
	r.notFoundHandler = handler
	return r
}

// OnError sets the error handler
func (r *Router) OnError(handler func(error) *dom.Element) *Router {
	r.errorHandler = handler
	return r
}

// compileRoute compiles route path to regex
func (r *Router) compileRoute(route *Route) {
	if route.Path == "" {
		return
	}

	pattern := route.Path
	paramNames := make([]string, 0)

	paramRegex := regexp.MustCompile(`:([a-zA-Z_][a-zA-Z0-9_]*)`)
	matches := paramRegex.FindAllStringSubmatch(pattern, -1)

	for _, match := range matches {
		paramNames = append(paramNames, match[1])
		pattern = strings.Replace(pattern, match[0], "([^/]+)", 1)
	}

	pattern = strings.Replace(pattern, "*", "(.*)", -1)
	pattern = "^" + pattern + "$"

	route.Regex = regexp.MustCompile(pattern)
	route.ParamNames = paramNames
}

// MatchRoute finds a matching route for the path. Exported for SSR use.
func (r *Router) MatchRoute(path string) (*Route, map[string]string) {
	for _, route := range r.routes {
		if route.Regex == nil {
			if route.Path == path {
				return route, make(map[string]string)
			}
			continue
		}

		matches := route.Regex.FindStringSubmatch(path)
		if matches != nil {
			params := make(map[string]string)
			for i, paramName := range route.ParamNames {
				if i+1 < len(matches) {
					params[paramName] = matches[i+1]
				}
			}
			return route, params
		}
	}

	return nil, nil
}

// GetNotFoundHandler returns the 404 handler
func (r *Router) GetNotFoundHandler() func() *dom.Element {
	return r.notFoundHandler
}

// GetErrorHandler returns the error handler
func (r *Router) GetErrorHandler() func(error) *dom.Element {
	return r.errorHandler
}

// GetRoutes returns the registered routes
func (r *Router) GetRoutes() []*Route {
	return r.routes
}

func (r *Router) Start() {
	fmt.Println("Router only available in WebAssembly build")
}

func (r *Router) Navigate(path string) error {
	return fmt.Errorf("routing only available in WebAssembly build")
}

func (r *Router) Push(path string) error {
	return fmt.Errorf("routing only available in WebAssembly build")
}

func (r *Router) Replace(path string) error {
	return fmt.Errorf("routing only available in WebAssembly build")
}

func (r *Router) Go(delta int)                                                  {}
func (r *Router) Back()                                                         {}
func (r *Router) Forward()                                                      {}
func (r *Router) GetCurrentRoute() *Route                                       { return nil }
func (r *Router) GetCurrentParams() map[string]string                           { return make(map[string]string) }
func (r *Router) GenerateURL(routeName string, params map[string]string) string { return "" }

func (l *LinkComponent) Render() *dom.Element {
	return dom.A(dom.Text(l.Text))
}

func RouterLink(router *Router, to, text string) *dom.Element {
	return dom.A(dom.Text(text))
}

func RouterLinkWithClass(router *Router, to, text, class string) *dom.Element {
	return dom.A(dom.Class(class), dom.Text(text))
}

func NewTransition() *Transition                                      { return &Transition{} }
func (t *Transition) AddHook(hook TransitionHook)                     {}
func (t *Transition) Execute(to *Route, from *Route, callback func()) { callback() }

var CommonGuards = &Guards{}

func (g *Guards) RequireAuth(isAuthenticated func() bool, redirectTo string) Guard {
	return func(to *Route, from *Route, params map[string]string) bool { return true }
}

func (g *Guards) RequireRole(hasRole func(role string) bool, role string) Guard {
	return func(to *Route, from *Route, params map[string]string) bool { return true }
}

func (g *Guards) ConfirmLeave(message string) Guard {
	return func(to *Route, from *Route, params map[string]string) bool { return true }
}

var DefaultRouter = NewRouter()

func AddRoute(path string, component func(params map[string]string) *dom.Element) {
	DefaultRouter.AddSimpleRoute(path, component)
}
func Navigate(path string) error { return fmt.Errorf("routing only available in WebAssembly build") }
func Push(path string) error     { return fmt.Errorf("routing only available in WebAssembly build") }
func Back()                      {}
func Forward()                   {}
func Start()                     { fmt.Println("Router only available in WebAssembly build") }
func CreateLink(to, text string) *dom.Element { return dom.A(dom.Text(text)) }
func CreateLinkWithClass(to, text, class string) *dom.Element {
	return dom.A(dom.Class(class), dom.Text(text))
}

// BuildLayoutChain wraps content in the route's layout and any parent layouts.
// It walks up the route hierarchy applying layouts from innermost to outermost.
func BuildLayoutChain(route *Route, content *dom.Element) *dom.Element {
	if route == nil || route.Layout == nil {
		return content
	}
	return route.Layout(content)
}

// RenderWithErrorBoundary renders the route component, catching panics and
// falling back to the error handler if one is provided.
func RenderWithErrorBoundary(route *Route, params map[string]string) (result *dom.Element, err error) {
	defer func() {
		if r := recover(); r != nil {
			var renderErr error
			switch v := r.(type) {
			case error:
				renderErr = v
			default:
				renderErr = fmt.Errorf("%v", v)
			}
			if route.ErrorHandler != nil {
				result = route.ErrorHandler(renderErr)
				err = nil
			} else {
				err = renderErr
			}
		}
	}()

	if route.Component == nil {
		return nil, fmt.Errorf("route %s has no component", route.Path)
	}

	result = route.Component(params)
	return result, nil
}

// RenderParallelSlots renders all parallel slot components for a route and
// returns them as a map of slot name to rendered element.
func RenderParallelSlots(route *Route, params map[string]string) map[string]*dom.Element {
	if route.ParallelSlots == nil {
		return nil
	}

	results := make(map[string]*dom.Element, len(route.ParallelSlots))
	for name, slotFn := range route.ParallelSlots {
		if slotFn != nil {
			results[name] = slotFn(params)
		}
	}
	return results
}
