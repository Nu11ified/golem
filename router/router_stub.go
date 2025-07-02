//go:build !js || !wasm

package router

import (
	"fmt"
	"regexp"

	"github.com/Nu11ified/golem/dom"
)

// Stub implementations for non-WASM builds
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
func (r *Router) AddRoute(route *Route) *Router        { return r }
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
