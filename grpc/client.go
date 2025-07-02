//go:build js && wasm

package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"syscall/js"
)

// Client provides seamless server function calling via gRPC-Web
type Client struct {
	baseURL string
	headers map[string]string
	timeout int // milliseconds
}

// NewClient creates a new gRPC-Web client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		headers: make(map[string]string),
		timeout: 30000, // 30 seconds default
	}
}

// SetHeader sets a header for all requests
func (c *Client) SetHeader(key, value string) {
	c.headers[key] = value
}

// SetTimeout sets the request timeout in milliseconds
func (c *Client) SetTimeout(timeout int) {
	c.timeout = timeout
}

// Call invokes a server function via gRPC-Web
func (c *Client) Call(ctx context.Context, serviceName, methodName string, req interface{}) (interface{}, error) {
	// Serialize request
	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create promise-based fetch call
	promise := c.createFetchPromise(serviceName, methodName, reqData)

	// Convert JS Promise to Go channel
	resultChan := make(chan fetchResult, 1)

	// Handle promise resolution
	promise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			response := args[0]
			resultChan <- fetchResult{response: response}
		}
		return nil
	}))

	// Handle promise rejection
	promise.Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			err := fmt.Errorf("fetch error: %s", args[0].String())
			resultChan <- fetchResult{error: err}
		}
		return nil
	}))

	// Wait for result or context cancellation
	select {
	case result := <-resultChan:
		if result.error != nil {
			return nil, result.error
		}
		return c.parseResponse(result.response)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

type fetchResult struct {
	response js.Value
	error    error
}

// createFetchPromise creates a JavaScript fetch promise for gRPC-Web
func (c *Client) createFetchPromise(serviceName, methodName string, reqData []byte) js.Value {
	url := fmt.Sprintf("%s/%s/%s", c.baseURL, serviceName, methodName)

	// Create fetch options
	options := js.Global().Get("Object").New()
	options.Set("method", "POST")
	options.Set("mode", "cors")

	// Set headers
	headers := js.Global().Get("Object").New()
	headers.Set("Content-Type", "application/json")
	headers.Set("Accept", "application/json")

	// Add custom headers
	for key, value := range c.headers {
		headers.Set(key, value)
	}

	options.Set("headers", headers)

	// Set body
	uint8Array := js.Global().Get("Uint8Array").New(len(reqData))
	js.CopyBytesToJS(uint8Array, reqData)
	options.Set("body", uint8Array)

	// Create fetch promise
	return js.Global().Call("fetch", url, options)
}

// parseResponse parses the fetch response
func (c *Client) parseResponse(response js.Value) (interface{}, error) {
	// Check if response is ok
	if !response.Get("ok").Bool() {
		status := response.Get("status").Int()
		statusText := response.Get("statusText").String()
		return nil, fmt.Errorf("HTTP %d: %s", status, statusText)
	}

	// Get response text
	textPromise := response.Call("text")

	// Convert promise to channel
	textChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	textPromise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			textChan <- args[0].String()
		}
		return nil
	}))

	textPromise.Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			errorChan <- fmt.Errorf("text parsing error: %s", args[0].String())
		}
		return nil
	}))

	// Wait for result
	select {
	case text := <-textChan:
		var result interface{}
		if err := json.Unmarshal([]byte(text), &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
		return result, nil
	case err := <-errorChan:
		return nil, err
	}
}

// ServerFunction provides a high-level interface for calling server functions
type ServerFunction struct {
	client      *Client
	serviceName string
	methodName  string
}

// NewServerFunction creates a new server function caller
func NewServerFunction(client *Client, serviceName, methodName string) *ServerFunction {
	return &ServerFunction{
		client:      client,
		serviceName: serviceName,
		methodName:  methodName,
	}
}

// Call invokes the server function with automatic type handling
func (sf *ServerFunction) Call(ctx context.Context, args ...interface{}) (interface{}, error) {
	// Handle different argument patterns
	var req interface{}

	switch len(args) {
	case 0:
		req = struct{}{}
	case 1:
		req = args[0]
	default:
		// Multiple arguments - wrap in struct
		req = map[string]interface{}{
			"args": args,
		}
	}

	return sf.client.Call(ctx, sf.serviceName, sf.methodName, req)
}

// CallWithResult invokes the server function and unmarshals result into target
func (sf *ServerFunction) CallWithResult(ctx context.Context, target interface{}, args ...interface{}) error {
	result, err := sf.Call(ctx, args...)
	if err != nil {
		return err
	}

	// Marshal and unmarshal to handle type conversion
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal into target: %w", err)
	}

	return nil
}

// Registry manages server function registrations
type Registry struct {
	functions map[string]*ServerFunction
	client    *Client
}

// NewRegistry creates a new function registry
func NewRegistry(client *Client) *Registry {
	return &Registry{
		functions: make(map[string]*ServerFunction),
		client:    client,
	}
}

// Register registers a server function
func (r *Registry) Register(name, serviceName, methodName string) {
	r.functions[name] = NewServerFunction(r.client, serviceName, methodName)
}

// Call calls a registered server function
func (r *Registry) Call(ctx context.Context, name string, args ...interface{}) (interface{}, error) {
	fn, exists := r.functions[name]
	if !exists {
		return nil, fmt.Errorf("server function %s not registered", name)
	}

	return fn.Call(ctx, args...)
}

// CallWithResult calls a registered server function with result unmarshaling
func (r *Registry) CallWithResult(ctx context.Context, name string, target interface{}, args ...interface{}) error {
	fn, exists := r.functions[name]
	if !exists {
		return fmt.Errorf("server function %s not registered", name)
	}

	return fn.CallWithResult(ctx, target, args...)
}

// Auto-registration helpers
func (r *Registry) RegisterServerPackage(packageName string, functions map[string]string) {
	for goFuncName, grpcMethodName := range functions {
		r.Register(goFuncName, packageName, grpcMethodName)
	}
}

// Type-safe server function calling with reflection
type TypedCall struct {
	registry *Registry
}

// NewTypedCall creates a new typed caller
func NewTypedCall(registry *Registry) *TypedCall {
	return &TypedCall{registry: registry}
}

// Call provides type-safe server function calling
func (tc *TypedCall) Call(ctx context.Context, fnName string, args interface{}, result interface{}) error {
	// Use reflection to validate types
	resultValue := reflect.ValueOf(result)

	if resultValue.Kind() != reflect.Ptr {
		return fmt.Errorf("result must be a pointer")
	}

	return tc.registry.CallWithResult(ctx, fnName, result, args)
}

// Streaming support for real-time updates
type Stream struct {
	client      *Client
	serviceName string
	methodName  string
	onMessage   func(interface{})
	onError     func(error)
	onClose     func()
}

// NewStream creates a new gRPC stream
func NewStream(client *Client, serviceName, methodName string) *Stream {
	return &Stream{
		client:      client,
		serviceName: serviceName,
		methodName:  methodName,
	}
}

// OnMessage sets the message handler
func (s *Stream) OnMessage(handler func(interface{})) *Stream {
	s.onMessage = handler
	return s
}

// OnError sets the error handler
func (s *Stream) OnError(handler func(error)) *Stream {
	s.onError = handler
	return s
}

// OnClose sets the close handler
func (s *Stream) OnClose(handler func()) *Stream {
	s.onClose = handler
	return s
}

// Start starts the stream
func (s *Stream) Start(ctx context.Context, req interface{}) error {
	// Implementation would use WebSocket or Server-Sent Events
	// for streaming gRPC communication
	return fmt.Errorf("streaming not yet implemented")
}

// Global client instance for easy access
var defaultClient *Client
var defaultRegistry *Registry

// SetDefaultClient sets the global default client
func SetDefaultClient(client *Client) {
	defaultClient = client
	defaultRegistry = NewRegistry(client)
}

// GetDefaultRegistry returns the default registry
func GetDefaultRegistry() *Registry {
	return defaultRegistry
}

// Call is a convenience function for calling server functions
func Call(ctx context.Context, name string, args ...interface{}) (interface{}, error) {
	if defaultRegistry == nil {
		return nil, fmt.Errorf("no default client configured")
	}
	return defaultRegistry.Call(ctx, name, args...)
}

// CallWithResult is a convenience function for calling server functions with result unmarshaling
func CallWithResult(ctx context.Context, name string, target interface{}, args ...interface{}) error {
	if defaultRegistry == nil {
		return fmt.Errorf("no default client configured")
	}
	return defaultRegistry.CallWithResult(ctx, name, target, args...)
}
