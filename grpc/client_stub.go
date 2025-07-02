//go:build !js || !wasm

package grpc

import (
	"context"
	"fmt"
)

// Stub implementations for non-WASM builds
type Client struct {
	baseURL string
	headers map[string]string
	timeout int
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		headers: make(map[string]string),
		timeout: 30000,
	}
}

func (c *Client) SetHeader(key, value string) {
	c.headers[key] = value
}

func (c *Client) SetTimeout(timeout int) {
	c.timeout = timeout
}

func (c *Client) Call(ctx context.Context, serviceName, methodName string, req interface{}) (interface{}, error) {
	return nil, fmt.Errorf("gRPC client only available in WebAssembly build")
}

type ServerFunction struct {
	client      *Client
	serviceName string
	methodName  string
}

func NewServerFunction(client *Client, serviceName, methodName string) *ServerFunction {
	return &ServerFunction{
		client:      client,
		serviceName: serviceName,
		methodName:  methodName,
	}
}

func (sf *ServerFunction) Call(ctx context.Context, args ...interface{}) (interface{}, error) {
	return nil, fmt.Errorf("server functions only available in WebAssembly build")
}

func (sf *ServerFunction) CallWithResult(ctx context.Context, target interface{}, args ...interface{}) error {
	return fmt.Errorf("server functions only available in WebAssembly build")
}

type Registry struct {
	functions map[string]*ServerFunction
	client    *Client
}

func NewRegistry(client *Client) *Registry {
	return &Registry{
		functions: make(map[string]*ServerFunction),
		client:    client,
	}
}

func (r *Registry) Register(name, serviceName, methodName string) {
	r.functions[name] = NewServerFunction(r.client, serviceName, methodName)
}

func (r *Registry) Call(ctx context.Context, name string, args ...interface{}) (interface{}, error) {
	return nil, fmt.Errorf("server functions only available in WebAssembly build")
}

func (r *Registry) CallWithResult(ctx context.Context, name string, target interface{}, args ...interface{}) error {
	return fmt.Errorf("server functions only available in WebAssembly build")
}

func (r *Registry) RegisterServerPackage(packageName string, functions map[string]string) {
	// No-op for stub
}

type TypedCall struct {
	registry *Registry
}

func NewTypedCall(registry *Registry) *TypedCall {
	return &TypedCall{registry: registry}
}

func (tc *TypedCall) Call(ctx context.Context, fnName string, args interface{}, result interface{}) error {
	return fmt.Errorf("typed calls only available in WebAssembly build")
}

type Stream struct {
	client      *Client
	serviceName string
	methodName  string
	onMessage   func(interface{})
	onError     func(error)
	onClose     func()
}

func NewStream(client *Client, serviceName, methodName string) *Stream {
	return &Stream{
		client:      client,
		serviceName: serviceName,
		methodName:  methodName,
	}
}

func (s *Stream) OnMessage(handler func(interface{})) *Stream {
	s.onMessage = handler
	return s
}

func (s *Stream) OnError(handler func(error)) *Stream {
	s.onError = handler
	return s
}

func (s *Stream) OnClose(handler func()) *Stream {
	s.onClose = handler
	return s
}

func (s *Stream) Start(ctx context.Context, req interface{}) error {
	return fmt.Errorf("streaming only available in WebAssembly build")
}

// Global stubs
var defaultClient *Client
var defaultRegistry *Registry

func SetDefaultClient(client *Client) {
	defaultClient = client
	defaultRegistry = NewRegistry(client)
}

func GetDefaultRegistry() *Registry {
	return defaultRegistry
}

func Call(ctx context.Context, name string, args ...interface{}) (interface{}, error) {
	return nil, fmt.Errorf("server functions only available in WebAssembly build")
}

func CallWithResult(ctx context.Context, name string, target interface{}, args ...interface{}) error {
	return fmt.Errorf("server functions only available in WebAssembly build")
}
