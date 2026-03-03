//go:build !js || !wasm

package router

import (
	"fmt"
	"regexp"

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
	Layout          func(children *dom.Element) *dom.Element
	ErrorHandler    func(err error) *dom.Element
	LoadingHandler  func() *dom.Element
	Template        func(children *dom.Element) *dom.Element
	ParentRoute     *Route
	ParallelSlots   map[string]*Route
	IsIntercepting  bool
	InterceptTarget string
}

type Guard func(to *Route, from *Route, params map[string]string) bool

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
		routes:        make([]*Route, 0),
		currentParams: make(map[string]string),
		beforeEach:    make([]Guard, 0),
		afterEach:     make([]func(*Route, *Route), 0),
		mode:          HashMode,
		container:     "#router-outlet",
	}
}

func (r *Router) SetMode(mode RouterMode) *Router      { return r }
func (r *Router) SetContainer(selector string) *Router { return r }
func (r *Router) SetBaseURL(baseURL string) *Router    { return r }
func (r *Router) AddRoute(route *Route) *Router {
	r.routes = append(r.routes, route)
	return r
}
func (r *Router) AddSimpleRoute(path string, component func(params map[string]string) *dom.Element) *Router {
	return r
}
func (r *Router) RouteWithName(name, path string, component func(params map[string]string) *dom.Element) *Router {
	return r
}
func (r *Router) RouteGroup(prefix string, guards []Guard, routes []*Route) *Router { return r }
func (r *Router) BeforeEach(guard Guard) *Router                                    { return r }
func (r *Router) AfterEach(hook func(*Route, *Route)) *Router                       { return r }
func (r *Router) NotFound(handler func() *dom.Element) *Router                      { return r }
func (r *Router) OnError(handler func(error) *dom.Element) *Router                  { return r }

// BuildLayoutChain walks the parent chain of a route and wraps the content
// in each Layout. The innermost layout wraps first, then the parent, etc.
// Result: rootLayout(parentLayout(childLayout(content)))
func (r *Router) BuildLayoutChain(route *Route, content *dom.Element) *dom.Element {
	if route == nil || content == nil {
		return content
	}

	result := content

	// Apply this route's layout first (innermost)
	if route.Layout != nil {
		result = route.Layout(result)
	}

	// Walk up the parent chain
	parent := route.ParentRoute
	for parent != nil {
		if parent.Layout != nil {
			result = parent.Layout(result)
		}
		parent = parent.ParentRoute
	}

	return result
}

// RenderWithErrorBoundary renders a route's component with panic recovery.
// If the component panics and the route has an ErrorHandler, the error
// handler is called with the recovered error to produce a fallback element.
// If there is no ErrorHandler, the panic propagates normally.
func (r *Router) RenderWithErrorBoundary(route *Route, params map[string]string) *dom.Element {
	if route.ErrorHandler == nil {
		return route.Component(params)
	}

	var result *dom.Element
	func() {
		defer func() {
			if err := recover(); err != nil {
				if e, ok := err.(error); ok {
					result = route.ErrorHandler(e)
				} else {
					result = route.ErrorHandler(fmt.Errorf("%v", err))
				}
			}
		}()
		result = route.Component(params)
	}()
	return result
}

// RenderParallelSlots renders each parallel slot independently with its own
// error boundary. The main route component (if any) is rendered as the
// "children" slot. Each slot gets its own error handling so a failure in
// one slot does not affect the others.
func (r *Router) RenderParallelSlots(route *Route, params map[string]string) map[string]*dom.Element {
	slots := make(map[string]*dom.Element)

	// Render main content as "children" slot
	if route.Component != nil {
		slots["children"] = r.RenderWithErrorBoundary(route, params)
	}

	// Render each parallel slot
	for name, slotRoute := range route.ParallelSlots {
		if slotRoute == nil || slotRoute.Component == nil {
			slots[name] = nil
			continue
		}
		slots[name] = r.renderSlotWithErrorBoundary(slotRoute, params)
	}

	return slots
}

// renderSlotWithErrorBoundary renders a slot route's component with panic
// recovery. Each slot uses its own ErrorHandler if available; otherwise
// the panic propagates.
func (r *Router) renderSlotWithErrorBoundary(slotRoute *Route, params map[string]string) *dom.Element {
	if slotRoute.ErrorHandler == nil {
		return slotRoute.Component(params)
	}

	var result *dom.Element
	func() {
		defer func() {
			if err := recover(); err != nil {
				if e, ok := err.(error); ok {
					result = slotRoute.ErrorHandler(e)
				} else {
					result = slotRoute.ErrorHandler(fmt.Errorf("%v", err))
				}
			}
		}()
		result = slotRoute.Component(params)
	}()
	return result
}

// AddRouteWithLayout adds a route and sets its parent for layout chaining.
func (r *Router) AddRouteWithLayout(route *Route, parent *Route) *Router {
	route.ParentRoute = parent
	return r.AddRoute(route)
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
