//go:build !js || !wasm

package router

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Nu11ified/golem/dom"
)

// Stub implementations for non-WASM builds
type Route struct {
	Path            string
	Component       func(params map[string]string) *dom.Element
	Guards          []Guard
	Children        []*Route
	Meta            map[string]interface{}
	Name            string
	Redirect        string
	Regex           *regexp.Regexp
	ParamNames      []string
	IsIntercepting  bool   // true if this route uses (.), (..), or (...) conventions
	InterceptTarget string // the target path this route intercepts (e.g., "/photos/:id")
}

type Guard func(to *Route, from *Route, params map[string]string) bool

type Router struct {
	routes             []*Route
	currentRoute       *Route
	currentParams      map[string]string
	beforeEach         []Guard
	afterEach          []func(*Route, *Route)
	notFoundHandler    func() *dom.Element
	errorHandler       func(error) *dom.Element
	baseURL            string
	mode               RouterMode
	container          string
	IsDirectNavigation bool // true when the page was loaded directly (not via client-side nav)
}

type RouterMode int

const (
	HashMode RouterMode = iota
	HistoryMode
)

type LinkComponent struct {
	To     string
	Class  string
	Text   string
	Router *Router
}

type TransitionHook func(to *Route, from *Route, next func())

type Transition struct {
	hooks []TransitionHook
}

type Guards struct{}

// Stub functions
func NewRouter() *Router {
	return &Router{
		routes:             make([]*Route, 0),
		currentParams:      make(map[string]string),
		beforeEach:         make([]Guard, 0),
		afterEach:          make([]func(*Route, *Route), 0),
		mode:               HashMode,
		container:          "#router-outlet",
		IsDirectNavigation: true, // initial page load is always direct
	}
}

func (r *Router) SetMode(mode RouterMode) *Router      { return r }
func (r *Router) SetContainer(selector string) *Router { return r }
func (r *Router) SetBaseURL(baseURL string) *Router    { return r }
func (r *Router) AddRoute(route *Route) *Router {
	r.compileRoute(route)
	r.routes = append(r.routes, route)
	return r
}
func (r *Router) AddSimpleRoute(path string, component func(params map[string]string) *dom.Element) *Router {
	return r.AddRoute(&Route{
		Path:      path,
		Component: component,
	})
}
func (r *Router) RouteWithName(name, path string, component func(params map[string]string) *dom.Element) *Router {
	return r
}
func (r *Router) RouteGroup(prefix string, guards []Guard, routes []*Route) *Router { return r }
func (r *Router) BeforeEach(guard Guard) *Router                                    { return r }
func (r *Router) AfterEach(hook func(*Route, *Route)) *Router                       { return r }
func (r *Router) NotFound(handler func() *dom.Element) *Router                      { return r }
func (r *Router) OnError(handler func(error) *dom.Element) *Router                  { return r }

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

// matchRoute finds a matching route for the path
func (r *Router) matchRoute(path string) (*Route, map[string]string) {
	for _, route := range r.routes {
		if route.IsIntercepting {
			continue // skip intercepting routes in normal matching
		}
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

// ResolveInterceptingRoute checks if the target path has an intercepting
// route registered and returns it if client-side navigation should use
// the intercepted version.
func (r *Router) ResolveInterceptingRoute(path string) *Route {
	for _, route := range r.routes {
		if !route.IsIntercepting {
			continue
		}
		if route.InterceptTarget == path {
			return route
		}
		if route.Regex != nil && route.Regex.MatchString(path) {
			return route
		}
	}
	return nil
}

func (r *Router) Start() {
	fmt.Println("Router only available in WebAssembly build")
}

func (r *Router) Navigate(path string) error {
	route, params := r.matchRoute(path)

	if route == nil {
		return fmt.Errorf("route not found: %s", path)
	}

	// For client-side navigation, check if an intercepting route should be used
	if !r.IsDirectNavigation {
		if interceptRoute := r.ResolveInterceptingRoute(path); interceptRoute != nil {
			route = interceptRoute
		}
	}

	r.currentRoute = route
	r.currentParams = params
	return nil
}

func (r *Router) Push(path string) error {
	r.IsDirectNavigation = false
	return r.Navigate(path)
}

func (r *Router) Replace(path string) error {
	return fmt.Errorf("routing only available in WebAssembly build")
}

func (r *Router) Go(delta int)                                                  {}
func (r *Router) Back()                                                         {}
func (r *Router) Forward()                                                      {}
func (r *Router) GetCurrentRoute() *Route                                       { return r.currentRoute }
func (r *Router) GetCurrentParams() map[string]string                           { return r.currentParams }
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

func AddRoute(path string, component func(params map[string]string) *dom.Element) {}
func Navigate(path string) error                                                  { return fmt.Errorf("routing only available in WebAssembly build") }
func Push(path string) error                                                      { return fmt.Errorf("routing only available in WebAssembly build") }
func Back()                                                                       {}
func Forward()                                                                    {}
func Start()                                                                      { fmt.Println("Router only available in WebAssembly build") }
func CreateLink(to, text string) *dom.Element                                     { return dom.A(dom.Text(text)) }
func CreateLinkWithClass(to, text, class string) *dom.Element {
	return dom.A(dom.Class(class), dom.Text(text))
}
