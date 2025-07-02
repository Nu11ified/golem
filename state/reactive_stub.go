//go:build !js || !wasm

package state

import (
	"fmt"
	"sync"

	"github.com/Nu11ified/golem/dom"
)

// Stub implementations for non-WASM builds
type Observable[T any] struct {
	value T
	mutex sync.RWMutex
}

type Observer[T any] func(newValue, oldValue T)

func NewObservable[T any](initialValue T) *Observable[T] {
	return &Observable[T]{value: initialValue}
}

func (o *Observable[T]) Get() T {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.value
}

func (o *Observable[T]) Set(newValue T) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.value = newValue
}

func (o *Observable[T]) Update(updateFn func(T) T) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.value = updateFn(o.value)
}

func (o *Observable[T]) Subscribe(observer Observer[T]) func() {
	return func() {} // No-op unsubscribe
}

func (o *Observable[T]) Map(mapFn func(T) interface{}) *Observable[interface{}] {
	return NewObservable[interface{}](mapFn(o.Get()))
}

func (o *Observable[T]) Filter(predicate func(T) bool) *Observable[T] {
	return NewObservable[T](o.value)
}

type Store struct {
	state      map[string]interface{}
	reducers   map[string]Reducer
	observers  map[string][]StoreObserver
	middleware []Middleware
	mutex      sync.RWMutex
}

type Action struct {
	Type    string
	Payload interface{}
}

type Reducer func(state interface{}, action Action) interface{}
type StoreObserver func(newState, oldState interface{})
type Middleware func(store *Store, action Action, next func(Action))

func NewStore() *Store {
	return &Store{
		state:      make(map[string]interface{}),
		reducers:   make(map[string]Reducer),
		observers:  make(map[string][]StoreObserver),
		middleware: make([]Middleware, 0),
	}
}

func (s *Store) AddReducer(key string, reducer Reducer, initialState interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.reducers[key] = reducer
	s.state[key] = initialState
}

func (s *Store) AddMiddleware(middleware Middleware) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.middleware = append(s.middleware, middleware)
}

func (s *Store) GetState(key string) interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.state[key]
}

func (s *Store) GetAllState() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	stateCopy := make(map[string]interface{})
	for k, v := range s.state {
		stateCopy[k] = v
	}
	return stateCopy
}

func (s *Store) Dispatch(action Action) {
	fmt.Printf("Store dispatch only available in WebAssembly build: %+v\n", action)
}

func (s *Store) Subscribe(key string, observer StoreObserver) func() {
	return func() {} // No-op unsubscribe
}

type Computed[T any] struct {
	value T
	mutex sync.RWMutex
}

func NewComputed[T any](computeFn func() T, deps ...interface{}) *Computed[T] {
	return &Computed[T]{value: computeFn()}
}

func (c *Computed[T]) Get() T {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.value
}

func (c *Computed[T]) Subscribe(observer Observer[T]) func() {
	return func() {} // No-op unsubscribe
}

type Component struct {
	render      func() *dom.Element
	state       map[string]interface{}
	observables map[string]interface{}
	element     *dom.Element
	mounted     bool
	mutex       sync.RWMutex
}

func NewComponent(renderFn func() *dom.Element) *Component {
	return &Component{
		render:      renderFn,
		state:       make(map[string]interface{}),
		observables: make(map[string]interface{}),
		mounted:     false,
	}
}

func (c *Component) UseState(key string, initialValue interface{}) *Observable[interface{}] {
	return NewObservable[interface{}](initialValue)
}

func (c *Component) UseStore(store *Store, key string) interface{} {
	return store.GetState(key)
}

func (c *Component) Mount(selector string) {
	fmt.Printf("Component mounting only available in WebAssembly build: %s\n", selector)
}

type Hooks struct {
	states    []interface{}
	effects   []Effect
	index     int
	component *Component
}

type Effect struct {
	fn      func()
	cleanup func()
	deps    []interface{}
}

func UseStateHook[T any](hooks *Hooks, initialValue T) (*Observable[T], func(T)) {
	observable := NewObservable[T](initialValue)
	setter := func(newValue T) {
		observable.Set(newValue)
	}
	return observable, setter
}

func UseEffect(hooks *Hooks, effectFn func(), deps []interface{}) {
	fmt.Println("UseEffect only available in WebAssembly build")
}

type Persistence struct{}

func NewPersistence() *Persistence {
	return &Persistence{}
}

func (p *Persistence) SaveState(key string, state interface{}) error {
	return fmt.Errorf("persistence only available in WebAssembly build")
}

func (p *Persistence) LoadState(key string, target interface{}) error {
	return fmt.Errorf("persistence only available in WebAssembly build")
}

func (p *Persistence) RemoveState(key string) {
	fmt.Printf("Persistence only available in WebAssembly build: %s\n", key)
}

type CommonMiddleware struct{}

var BuiltinMiddleware = &CommonMiddleware{}

func (m *CommonMiddleware) Logger(store *Store, action Action, next func(Action)) {
	fmt.Printf("Logger middleware only available in WebAssembly build: %+v\n", action)
	next(action)
}

func (m *CommonMiddleware) Persistence(persistence *Persistence, keys []string) Middleware {
	return func(store *Store, action Action, next func(Action)) {
		next(action)
	}
}

func (m *CommonMiddleware) DevTools(store *Store, action Action, next func(Action)) {
	next(action)
}

var GlobalStore = NewStore()

func CreateObservable[T any](value T) *Observable[T] {
	return NewObservable[T](value)
}

func CreateStore() *Store {
	return NewStore()
}

func CreateComponent(renderFn func() *dom.Element) *Component {
	return NewComponent(renderFn)
}

func CreatePersistence() *Persistence {
	return NewPersistence()
}
