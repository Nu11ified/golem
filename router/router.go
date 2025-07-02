//go:build js && wasm

package router

import (
	"fmt"
	"regexp"
	"strings"
	"syscall/js"

	"github.com/Nu11ified/golem/dom"
)

// Route represents a single route
type Route struct {
	Path       string
	Component  func(params map[string]string) *dom.Element
	Guards     []Guard
	Children   []*Route
	Meta       map[string]interface{}
	Name       string
	Redirect   string
	Regex      *regexp.Regexp
	ParamNames []string
}

// Guard represents a route guard
type Guard func(to *Route, from *Route, params map[string]string) bool

// Router manages client-side routing
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
	container       string // CSS selector for router outlet
}

// RouterMode defines routing modes
type RouterMode int

const (
	HashMode    RouterMode = iota // #/path
	HistoryMode                   // /path (requires server support)
)

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

// SetMode sets the router mode
func (r *Router) SetMode(mode RouterMode) *Router {
	r.mode = mode
	return r
}

// SetContainer sets the router outlet container
func (r *Router) SetContainer(selector string) *Router {
	r.container = selector
	return r
}

// SetBaseURL sets the base URL for history mode
func (r *Router) SetBaseURL(baseURL string) *Router {
	r.baseURL = strings.TrimSuffix(baseURL, "/")
	return r
}

// AddRoute adds a route to the router
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

// compileRoute compiles route path to regex
func (r *Router) compileRoute(route *Route) {
	if route.Path == "" {
		return
	}

	// Handle wildcards and parameters
	pattern := route.Path
	paramNames := make([]string, 0)

	// Replace parameters like :id with regex groups
	paramRegex := regexp.MustCompile(`:([a-zA-Z_][a-zA-Z0-9_]*)`)
	matches := paramRegex.FindAllStringSubmatch(pattern, -1)

	for _, match := range matches {
		paramNames = append(paramNames, match[1])
		pattern = strings.Replace(pattern, match[0], "([^/]+)", 1)
	}

	// Handle wildcards
	pattern = strings.Replace(pattern, "*", "(.*)", -1)

	// Anchor pattern
	pattern = "^" + pattern + "$"

	route.Regex = regexp.MustCompile(pattern)
	route.ParamNames = paramNames
}

// BeforeEach adds a global before guard
func (r *Router) BeforeEach(guard Guard) *Router {
	r.beforeEach = append(r.beforeEach, guard)
	return r
}

// AfterEach adds a global after hook
func (r *Router) AfterEach(hook func(*Route, *Route)) *Router {
	r.afterEach = append(r.afterEach, hook)
	return r
}

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

// Start initializes the router
func (r *Router) Start() {
	// Listen for browser navigation events
	r.setupEventListeners()

	// Handle initial route
	r.handleCurrentLocation()
}

// setupEventListeners sets up browser event listeners
func (r *Router) setupEventListeners() {
	window := js.Global().Get("window")

	if r.mode == HistoryMode {
		// Listen for popstate events (back/forward buttons)
		popstateHandler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			r.handleCurrentLocation()
			return nil
		})
		window.Call("addEventListener", "popstate", popstateHandler)
	} else {
		// Listen for hashchange events
		hashchangeHandler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			r.handleCurrentLocation()
			return nil
		})
		window.Call("addEventListener", "hashchange", hashchangeHandler)
	}
}

// getCurrentPath gets the current path from the URL
func (r *Router) getCurrentPath() string {
	location := js.Global().Get("location")

	if r.mode == HistoryMode {
		pathname := location.Get("pathname").String()
		if r.baseURL != "" {
			pathname = strings.TrimPrefix(pathname, r.baseURL)
		}
		return pathname
	} else {
		hash := location.Get("hash").String()
		if hash == "" {
			return "/"
		}
		return strings.TrimPrefix(hash, "#")
	}
}

// handleCurrentLocation handles the current location
func (r *Router) handleCurrentLocation() {
	path := r.getCurrentPath()
	r.Navigate(path)
}

// Navigate navigates to a path
func (r *Router) Navigate(path string) error {
	route, params := r.matchRoute(path)

	if route == nil {
		if r.notFoundHandler != nil {
			r.renderComponent(r.notFoundHandler())
			return nil
		}
		return fmt.Errorf("route not found: %s", path)
	}

	// Check guards
	if !r.checkGuards(route, r.currentRoute, params) {
		return fmt.Errorf("navigation blocked by guard")
	}

	// Handle redirect
	if route.Redirect != "" {
		return r.Navigate(route.Redirect)
	}

	// Update browser URL
	r.updateURL(path)

	// Update current route
	previousRoute := r.currentRoute
	r.currentRoute = route
	r.currentParams = params

	// Render component
	if route.Component != nil {
		component := route.Component(params)
		r.renderComponent(component)
	}

	// Run after hooks
	for _, hook := range r.afterEach {
		hook(route, previousRoute)
	}

	return nil
}

// matchRoute finds a matching route for the path
func (r *Router) matchRoute(path string) (*Route, map[string]string) {
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

// checkGuards runs all guards for a route
func (r *Router) checkGuards(to *Route, from *Route, params map[string]string) bool {
	// Global before guards
	for _, guard := range r.beforeEach {
		if !guard(to, from, params) {
			return false
		}
	}

	// Route-specific guards
	for _, guard := range to.Guards {
		if !guard(to, from, params) {
			return false
		}
	}

	return true
}

// updateURL updates the browser URL
func (r *Router) updateURL(path string) {
	history := js.Global().Get("history")

	if r.mode == HistoryMode {
		url := r.baseURL + path
		history.Call("pushState", nil, "", url)
	} else {
		js.Global().Get("location").Set("hash", "#"+path)
	}
}

// renderComponent renders a component in the router outlet
func (r *Router) renderComponent(component *dom.Element) {
	if component == nil {
		return
	}

	// Find router outlet
	doc := js.Global().Get("document")
	outlet := doc.Call("querySelector", r.container)

	if outlet.IsNull() {
		fmt.Printf("Router outlet not found: %s\n", r.container)
		return
	}

	// Clear outlet
	outlet.Set("innerHTML", "")

	// Render component
	renderedElement := component.Render()
	outlet.Call("appendChild", renderedElement)
}

// Push navigates to a new route
func (r *Router) Push(path string) error {
	return r.Navigate(path)
}

// Replace replaces the current route
func (r *Router) Replace(path string) error {
	route, params := r.matchRoute(path)

	if route == nil {
		return fmt.Errorf("route not found: %s", path)
	}

	// Check guards
	if !r.checkGuards(route, r.currentRoute, params) {
		return fmt.Errorf("navigation blocked by guard")
	}

	// Update browser URL (replace instead of push)
	history := js.Global().Get("history")
	if r.mode == HistoryMode {
		url := r.baseURL + path
		history.Call("replaceState", nil, "", url)
	} else {
		js.Global().Get("location").Call("replace", "#"+path)
	}

	// Update current route
	r.currentRoute = route
	r.currentParams = params

	// Render component
	if route.Component != nil {
		component := route.Component(params)
		r.renderComponent(component)
	}

	return nil
}

// Go navigates back/forward in history
func (r *Router) Go(delta int) {
	history := js.Global().Get("history")
	history.Call("go", delta)
}

// Back navigates back
func (r *Router) Back() {
	r.Go(-1)
}

// Forward navigates forward
func (r *Router) Forward() {
	r.Go(1)
}

// GetCurrentRoute returns the current route
func (r *Router) GetCurrentRoute() *Route {
	return r.currentRoute
}

// GetCurrentParams returns the current route parameters
func (r *Router) GetCurrentParams() map[string]string {
	return r.currentParams
}

// GenerateURL generates a URL for a named route
func (r *Router) GenerateURL(routeName string, params map[string]string) string {
	for _, route := range r.routes {
		if route.Name == routeName {
			path := route.Path
			for paramName, paramValue := range params {
				path = strings.Replace(path, ":"+paramName, paramValue, -1)
			}
			return path
		}
	}
	return ""
}

// LinkComponent for navigation
type LinkComponent struct {
	To     string
	Class  string
	Text   string
	Router *Router
}

// Render renders a navigation link
func (l *LinkComponent) Render() *dom.Element {
	return dom.A(
		dom.Class(l.Class),
		dom.Text(l.Text),
		dom.OnClick(func() {
			l.Router.Push(l.To)
		}),
	)
}

// RouterLink creates a navigation link
func RouterLink(router *Router, to, text string) *dom.Element {
	link := &LinkComponent{
		To:     to,
		Text:   text,
		Router: router,
	}
	return link.Render()
}

// RouterLinkWithClass creates a navigation link with CSS class
func RouterLinkWithClass(router *Router, to, text, class string) *dom.Element {
	link := &LinkComponent{
		To:     to,
		Text:   text,
		Class:  class,
		Router: router,
	}
	return link.Render()
}

// Route transition hooks
type TransitionHook func(to *Route, from *Route, next func())

// Transition manages route transitions
type Transition struct {
	hooks []TransitionHook
}

// NewTransition creates a new transition manager
func NewTransition() *Transition {
	return &Transition{
		hooks: make([]TransitionHook, 0),
	}
}

// AddHook adds a transition hook
func (t *Transition) AddHook(hook TransitionHook) {
	t.hooks = append(t.hooks, hook)
}

// Execute executes all transition hooks
func (t *Transition) Execute(to *Route, from *Route, callback func()) {
	if len(t.hooks) == 0 {
		callback()
		return
	}

	index := 0

	var next func()
	next = func() {
		if index >= len(t.hooks) {
			callback()
			return
		}

		hook := t.hooks[index]
		index++
		hook(to, from, next)
	}

	next()
}

// Common route guards
type Guards struct{}

var CommonGuards = &Guards{}

// RequireAuth creates an authentication guard
func (g *Guards) RequireAuth(isAuthenticated func() bool, redirectTo string) Guard {
	return func(to *Route, from *Route, params map[string]string) bool {
		if !isAuthenticated() {
			// In a real app, you'd redirect here
			fmt.Printf("Authentication required for route: %s\n", to.Path)
			return false
		}
		return true
	}
}

// RequireRole creates a role-based guard
func (g *Guards) RequireRole(hasRole func(role string) bool, role string) Guard {
	return func(to *Route, from *Route, params map[string]string) bool {
		if !hasRole(role) {
			fmt.Printf("Role %s required for route: %s\n", role, to.Path)
			return false
		}
		return true
	}
}

// ConfirmLeave creates a confirmation guard for leaving a route
func (g *Guards) ConfirmLeave(message string) Guard {
	return func(to *Route, from *Route, params map[string]string) bool {
		// In a real app, you'd show a confirmation dialog
		fmt.Printf("Confirm leave: %s\n", message)
		return true // For now, always allow
	}
}

// Global router instance
var DefaultRouter = NewRouter()

// Convenience functions for the default router
func AddRoute(path string, component func(params map[string]string) *dom.Element) {
	DefaultRouter.AddSimpleRoute(path, component)
}

func Navigate(path string) error {
	return DefaultRouter.Navigate(path)
}

func Push(path string) error {
	return DefaultRouter.Push(path)
}

func Back() {
	DefaultRouter.Back()
}

func Forward() {
	DefaultRouter.Forward()
}

func Start() {
	DefaultRouter.Start()
}

func CreateLink(to, text string) *dom.Element {
	return RouterLink(DefaultRouter, to, text)
}

func CreateLinkWithClass(to, text, class string) *dom.Element {
	return RouterLinkWithClass(DefaultRouter, to, text, class)
}
