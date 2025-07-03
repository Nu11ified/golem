//go:build js && wasm

package state

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"syscall/js"

	"github.com/Nu11ified/golem/dom"
)

// Observable represents a reactive value
type Observable[T any] struct {
	value     T
	observers []Observer[T]
	mutex     sync.RWMutex
}

// Observer represents a function that gets called when value changes
type Observer[T any] func(newValue, oldValue T)

// NewObservable creates a new observable value
func NewObservable[T any](initialValue T) *Observable[T] {
	return &Observable[T]{
		value:     initialValue,
		observers: make([]Observer[T], 0),
	}
}

// Get returns the current value
func (o *Observable[T]) Get() T {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.value
}

// Set updates the value and notifies observers
func (o *Observable[T]) Set(newValue T) {
	o.mutex.Lock()
	oldValue := o.value
	o.value = newValue
	observers := make([]Observer[T], len(o.observers))
	copy(observers, o.observers)
	o.mutex.Unlock()

	// Notify observers outside the lock to prevent deadlocks
	for _, observer := range observers {
		observer(newValue, oldValue)
	}
}

// Update modifies the value using a function
func (o *Observable[T]) Update(updateFn func(T) T) {
	o.mutex.Lock()
	oldValue := o.value
	newValue := updateFn(oldValue)
	o.value = newValue
	observers := make([]Observer[T], len(o.observers))
	copy(observers, o.observers)
	o.mutex.Unlock()

	for _, observer := range observers {
		observer(newValue, oldValue)
	}
}

// Subscribe adds an observer
func (o *Observable[T]) Subscribe(observer Observer[T]) func() {
	o.mutex.Lock()
	o.observers = append(o.observers, observer)
	index := len(o.observers) - 1
	o.mutex.Unlock()

	// Return unsubscribe function
	return func() {
		o.mutex.Lock()
		defer o.mutex.Unlock()
		if index < len(o.observers) {
			o.observers = append(o.observers[:index], o.observers[index+1:]...)
		}
	}
}

// Map creates a new observable that transforms this one
func (o *Observable[T]) Map(mapFn func(T) interface{}) *Observable[interface{}] {
	mapped := NewObservable(mapFn(o.Get()))

	o.Subscribe(func(newValue, oldValue T) {
		mapped.Set(mapFn(newValue))
	})

	return mapped
}

// Filter creates a new observable that only emits when predicate is true
func (o *Observable[T]) Filter(predicate func(T) bool) *Observable[T] {
	var zero T
	filtered := NewObservable(zero)

	o.Subscribe(func(newValue, oldValue T) {
		if predicate(newValue) {
			filtered.Set(newValue)
		}
	})

	return filtered
}

// Store represents a centralized state store
type Store struct {
	state      map[string]interface{}
	reducers   map[string]Reducer
	observers  map[string][]StoreObserver
	middleware []Middleware
	mutex      sync.RWMutex
}

// Action represents a state change action
type Action struct {
	Type    string
	Payload interface{}
}

// Reducer represents a function that updates state based on actions
type Reducer func(state interface{}, action Action) interface{}

// StoreObserver represents a function that gets called when store state changes
type StoreObserver func(newState, oldState interface{})

// Middleware represents middleware that can intercept actions
type Middleware func(store *Store, action Action, next func(Action))

// NewStore creates a new state store
func NewStore() *Store {
	return &Store{
		state:      make(map[string]interface{}),
		reducers:   make(map[string]Reducer),
		observers:  make(map[string][]StoreObserver),
		middleware: make([]Middleware, 0),
	}
}

// AddReducer adds a reducer for a specific state key
func (s *Store) AddReducer(key string, reducer Reducer, initialState interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.reducers[key] = reducer
	s.state[key] = initialState
}

// AddMiddleware adds middleware to the store
func (s *Store) AddMiddleware(middleware Middleware) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.middleware = append(s.middleware, middleware)
}

// GetState returns the current state for a key
func (s *Store) GetState(key string) interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.state[key]
}

// GetAllState returns the entire state
func (s *Store) GetAllState() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	stateCopy := make(map[string]interface{})
	for k, v := range s.state {
		stateCopy[k] = v
	}
	return stateCopy
}

// Dispatch dispatches an action through middleware and reducers
func (s *Store) Dispatch(action Action) {
	if len(s.middleware) == 0 {
		s.dispatchToReducers(action)
		return
	}

	// Execute middleware chain
	index := 0
	var next func(Action)
	next = func(a Action) {
		if index >= len(s.middleware) {
			s.dispatchToReducers(a)
			return
		}

		middleware := s.middleware[index]
		index++
		middleware(s, a, next)
	}

	next(action)
}

// dispatchToReducers applies action to all reducers
func (s *Store) dispatchToReducers(action Action) {
	s.mutex.Lock()

	oldState := make(map[string]interface{})
	for k, v := range s.state {
		oldState[k] = v
	}

	// Apply reducers
	for key, reducer := range s.reducers {
		if currentState, exists := s.state[key]; exists {
			newState := reducer(currentState, action)
			s.state[key] = newState
		}
	}

	// Get observers to notify
	observersToNotify := make(map[string][]StoreObserver)
	for key, observers := range s.observers {
		observersToNotify[key] = make([]StoreObserver, len(observers))
		copy(observersToNotify[key], observers)
	}

	s.mutex.Unlock()

	// Notify observers
	for key, observers := range observersToNotify {
		newState := s.GetState(key)
		oldStateValue := oldState[key]

		for _, observer := range observers {
			observer(newState, oldStateValue)
		}
	}
}

// Subscribe subscribes to state changes for a specific key
func (s *Store) Subscribe(key string, observer StoreObserver) func() {
	s.mutex.Lock()

	if s.observers[key] == nil {
		s.observers[key] = make([]StoreObserver, 0)
	}

	s.observers[key] = append(s.observers[key], observer)
	index := len(s.observers[key]) - 1

	s.mutex.Unlock()

	// Return unsubscribe function
	return func() {
		s.mutex.Lock()
		defer s.mutex.Unlock()

		if observers, exists := s.observers[key]; exists && index < len(observers) {
			s.observers[key] = append(observers[:index], observers[index+1:]...)
		}
	}
}

// Computed represents a computed value that depends on other observables
type Computed[T any] struct {
	computeFn func() T
	value     T
	observers []Observer[T]
	deps      []interface{} // Dependencies
	mutex     sync.RWMutex
}

// NewComputed creates a new computed observable
func NewComputed[T any](computeFn func() T, deps ...interface{}) *Computed[T] {
	computed := &Computed[T]{
		computeFn: computeFn,
		value:     computeFn(),
		observers: make([]Observer[T], 0),
		deps:      deps,
	}

	// Subscribe to dependencies
	for _, dep := range deps {
		switch d := dep.(type) {
		case *Observable[interface{}]:
			d.Subscribe(func(newValue, oldValue interface{}) {
				computed.recompute()
			})
		case *Store:
			// For stores, we'd need to know which keys to watch
			// This is a simplified implementation
		}
	}

	return computed
}

// recompute recalculates the value
func (c *Computed[T]) recompute() {
	c.mutex.Lock()
	oldValue := c.value
	newValue := c.computeFn()
	c.value = newValue
	observers := make([]Observer[T], len(c.observers))
	copy(observers, c.observers)
	c.mutex.Unlock()

	for _, observer := range observers {
		observer(newValue, oldValue)
	}
}

// Get returns the current computed value
func (c *Computed[T]) Get() T {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.value
}

// Subscribe adds an observer to the computed value
func (c *Computed[T]) Subscribe(observer Observer[T]) func() {
	c.mutex.Lock()
	c.observers = append(c.observers, observer)
	index := len(c.observers) - 1
	c.mutex.Unlock()

	return func() {
		c.mutex.Lock()
		defer c.mutex.Unlock()
		if index < len(c.observers) {
			c.observers = append(c.observers[:index], c.observers[index+1:]...)
		}
	}
}

// Component represents a reactive component
type Component struct {
	render      func() *dom.Element
	state       map[string]interface{}
	observables map[string]interface{}
	element     *dom.Element
	mounted     bool
	mutex       sync.RWMutex
}

// NewComponent creates a new reactive component
func NewComponent(renderFn func() *dom.Element) *Component {
	return &Component{
		render:      renderFn,
		state:       make(map[string]interface{}),
		observables: make(map[string]interface{}),
		mounted:     false,
	}
}

// UseState creates a state variable for the component
func (c *Component) UseState(key string, initialValue interface{}) *Observable[interface{}] {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if obs, exists := c.observables[key]; exists {
		return obs.(*Observable[interface{}])
	}

	observable := NewObservable(initialValue)
	c.observables[key] = observable

	// Subscribe to re-render on changes
	observable.Subscribe(func(newValue, oldValue interface{}) {
		if c.mounted {
			c.rerender()
		}
	})

	return observable
}

// UseStore connects the component to a store
func (c *Component) UseStore(store *Store, key string) interface{} {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	currentState := store.GetState(key)

	// Subscribe to store changes
	store.Subscribe(key, func(newState, oldState interface{}) {
		if c.mounted {
			c.rerender()
		}
	})

	return currentState
}

// Mount mounts the component to the DOM
func (c *Component) Mount(selector string) {
	c.mutex.Lock()
	c.mounted = true
	c.mutex.Unlock()

	c.rerender()
}

// rerender re-renders the component
func (c *Component) rerender() {
	if !c.mounted {
		return
	}

	newElement := c.render()
	if c.element != nil {
		// In a real implementation, we'd use the virtual DOM diffing here
		// For now, just replace the entire element
	}
	c.element = newElement
}

// Hooks for functional components
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

// UseStateHook creates a state hook
func UseStateHook[T any](hooks *Hooks, initialValue T) (*Observable[T], func(T)) {
	if hooks.index >= len(hooks.states) {
		observable := NewObservable(initialValue)
		hooks.states = append(hooks.states, observable)
	}

	observable := hooks.states[hooks.index].(*Observable[T])
	hooks.index++

	setter := func(newValue T) {
		observable.Set(newValue)
	}

	return observable, setter
}

// UseEffect adds an effect hook
func UseEffect(hooks *Hooks, effectFn func(), deps []interface{}) {
	if hooks.index >= len(hooks.effects) {
		effect := Effect{
			fn:   effectFn,
			deps: deps,
		}
		hooks.effects = append(hooks.effects, effect)

		// Run effect immediately
		effectFn()
	} else {
		effect := &hooks.effects[hooks.index]

		// Check if dependencies changed
		depsChanged := false
		if len(effect.deps) != len(deps) {
			depsChanged = true
		} else {
			for i, dep := range deps {
				if !reflect.DeepEqual(dep, effect.deps[i]) {
					depsChanged = true
					break
				}
			}
		}

		if depsChanged {
			// Cleanup previous effect
			if effect.cleanup != nil {
				effect.cleanup()
			}

			// Run new effect
			effect.fn = effectFn
			effect.deps = deps
			effectFn()
		}
	}

	hooks.index++
}

// Persistence layer
type Persistence struct {
	storage js.Value
}

// NewPersistence creates a new persistence layer
func NewPersistence() *Persistence {
	return &Persistence{
		storage: js.Global().Get("localStorage"),
	}
}

// SaveState saves state to localStorage
func (p *Persistence) SaveState(key string, state interface{}) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	p.storage.Call("setItem", key, string(data))
	return nil
}

// LoadState loads state from localStorage
func (p *Persistence) LoadState(key string, target interface{}) error {
	item := p.storage.Call("getItem", key)
	if item.IsNull() {
		return fmt.Errorf("no state found for key: %s", key)
	}

	data := item.String()
	return json.Unmarshal([]byte(data), target)
}

// RemoveState removes state from localStorage
func (p *Persistence) RemoveState(key string) {
	p.storage.Call("removeItem", key)
}

// Common middleware
type CommonMiddleware struct{}

var BuiltinMiddleware = &CommonMiddleware{}

// Logger middleware logs all actions
func (m *CommonMiddleware) Logger(store *Store, action Action, next func(Action)) {
	fmt.Printf("Action: %+v\n", action)
	oldState := store.GetAllState()

	next(action)

	newState := store.GetAllState()
	fmt.Printf("State changed from %+v to %+v\n", oldState, newState)
}

// Persistence middleware automatically saves state
func (m *CommonMiddleware) Persistence(persistence *Persistence, keys []string) Middleware {
	return func(store *Store, action Action, next func(Action)) {
		next(action)

		// Save specified keys to persistence
		for _, key := range keys {
			state := store.GetState(key)
			if state != nil {
				persistence.SaveState(key, state)
			}
		}
	}
}

// DevTools middleware for development
func (m *CommonMiddleware) DevTools(store *Store, action Action, next func(Action)) {
	// In development, we could send state to browser dev tools
	next(action)
}

// Global store instance
var GlobalStore = NewStore()

// Convenience functions
func CreateObservable[T any](value T) *Observable[T] {
	return NewObservable(value)
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

// ReactiveState represents a reactive state manager for complex state objects
type ReactiveState struct {
	value     interface{}
	observers []func(interface{})
	mutex     sync.RWMutex
}

// NewReactiveState creates a new reactive state manager
func NewReactiveState(initialValue interface{}) *ReactiveState {
	return &ReactiveState{
		value:     initialValue,
		observers: make([]func(interface{}), 0),
	}
}

// Get returns the current state value
func (rs *ReactiveState) Get() interface{} {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()
	return rs.value
}

// Update modifies the state using an updater function and notifies observers
func (rs *ReactiveState) Update(updater func(interface{}) interface{}) {
	rs.mutex.Lock()
	newValue := updater(rs.value)
	rs.value = newValue
	observers := make([]func(interface{}), len(rs.observers))
	copy(observers, rs.observers)
	rs.mutex.Unlock()

	fmt.Printf("🔄 ReactiveState.Update: state changed, notifying %d observers\n", len(observers))

	// Notify observers outside the lock
	for i, observer := range observers {
		fmt.Printf("  📢 Notifying observer %d\n", i)
		observer(newValue)
	}
}

// Subscribe adds an observer that gets called when state changes
func (rs *ReactiveState) Subscribe(observer func(interface{})) func() {
	rs.mutex.Lock()
	rs.observers = append(rs.observers, observer)
	index := len(rs.observers) - 1
	rs.mutex.Unlock()

	// Return unsubscribe function
	return func() {
		rs.mutex.Lock()
		defer rs.mutex.Unlock()
		if index < len(rs.observers) {
			rs.observers = append(rs.observers[:index], rs.observers[index+1:]...)
		}
	}
}

// WithState creates a reactive DOM element that updates when state changes
func (rs *ReactiveState) WithState(renderFn func(interface{}) *dom.Element) *dom.Element {
	// Initial render
	element := renderFn(rs.Get())
	fmt.Printf("🎨 ReactiveState.WithState: Initial render complete\n")

	// Subscribe to state changes and re-render
	rs.Subscribe(func(newState interface{}) {
		fmt.Printf("🎨 ReactiveState.WithState: State changed, triggering re-render\n")
		newElement := renderFn(newState)

		// Ensure both elements are rendered
		if element.JSElement.IsUndefined() {
			fmt.Printf("  🔧 Initial element not rendered, rendering now\n")
			element.Render()
		}

		renderedNewElement := newElement.Render()
		fmt.Printf("  🔧 New element rendered\n")

		// Replace the old element with the new one in the DOM
		if !element.JSElement.IsUndefined() {
			parent := element.JSElement.Get("parentNode")
			if !parent.IsUndefined() && !parent.IsNull() {
				fmt.Printf("  🔄 Replacing DOM element\n")
				parent.Call("replaceChild", renderedNewElement, element.JSElement)

				// Update the element reference to point to the new DOM node
				element.JSElement = renderedNewElement
				element.Props = newElement.Props
				element.Children = newElement.Children
				element.Type = newElement.Type
				element.EventHandlers = newElement.EventHandlers
				fmt.Printf("  ✅ DOM element replaced successfully\n")
			} else {
				fmt.Printf("  ❌ Parent element not found in DOM\n")
			}
		} else {
			fmt.Printf("  ❌ Original element JSElement is undefined\n")
		}
	})

	return element
}
