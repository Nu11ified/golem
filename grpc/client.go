//go:build js && wasm

package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"syscall/js"
	"time"
)

// Client provides seamless server function calling from frontend
type Client struct {
	baseURL string
	timeout time.Duration
}

// NewClient creates a new client for calling server functions
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		timeout: 30 * time.Second,
	}
}

// SetTimeout sets the request timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// Call invokes a server function with automatic argument marshaling
func (c *Client) Call(ctx context.Context, serviceName, functionName string, args ...interface{}) (interface{}, error) {
	// Create the request payload
	requestData := map[string]interface{}{
		"functionName": functionName,
		"serviceName":  serviceName,
		"args":         args,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make the HTTP request using fetch
	return c.makeRequest(ctx, jsonData)
}

// makeRequest performs the actual HTTP request using JavaScript fetch
func (c *Client) makeRequest(ctx context.Context, jsonData []byte) (interface{}, error) {
	// Create a promise-based approach
	resultChan := make(chan fetchResult, 1)

	// Create fetch options
	options := js.Global().Get("Object").New()
	options.Set("method", "POST")
	options.Set("mode", "cors")

	// Set headers
	headers := js.Global().Get("Object").New()
	headers.Set("Content-Type", "application/json")
	headers.Set("Accept", "application/json")
	options.Set("headers", headers)

	// Set body
	options.Set("body", string(jsonData))

	// Build the URL
	url := fmt.Sprintf("%s/api/functions", c.baseURL)

	// Debug logging
	fmt.Printf("🌐 gRPC Client Debug:\n")
	fmt.Printf("  baseURL: '%s'\n", c.baseURL)
	fmt.Printf("  Final URL: '%s'\n", url)
	fmt.Printf("  Request body: %s\n", string(jsonData))

	// Make the fetch call
	promise := js.Global().Call("fetch", url, options)

	// Handle promise resolution
	var thenFunc js.Func
	thenFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer thenFunc.Release() // Release after callback completes
		if len(args) > 0 {
			response := args[0]
			fmt.Printf("📥 HTTP Response: status=%d, ok=%t\n", response.Get("status").Int(), response.Get("ok").Bool())
			// Process the response synchronously to avoid race conditions
			c.processResponse(response, resultChan)
		}
		return nil
	})

	// Handle promise rejection
	var catchFunc js.Func
	catchFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer catchFunc.Release() // Release after callback completes
		if len(args) > 0 {
			err := fmt.Errorf("fetch error: %s", args[0].String())
			fmt.Printf("❌ Fetch error: %v\n", err)
			resultChan <- fetchResult{error: err}
		}
		return nil
	})

	promise.Call("then", thenFunc).Call("catch", catchFunc)

	// Wait for result or context cancellation
	select {
	case result := <-resultChan:
		if result.error != nil {
			fmt.Printf("❌ Final error: %v\n", result.error)
			return nil, result.error
		}
		fmt.Printf("✅ Final result: %+v\n", result.data)
		return result.data, nil
	case <-ctx.Done():
		fmt.Printf("❌ Context cancelled: %v\n", ctx.Err())
		return nil, ctx.Err()
	case <-time.After(c.timeout):
		fmt.Printf("❌ Request timeout after %v\n", c.timeout)
		return nil, fmt.Errorf("request timeout after %v", c.timeout)
	}
}

type fetchResult struct {
	data  interface{}
	error error
}

// processResponse processes the fetch response synchronously
func (c *Client) processResponse(response js.Value, resultChan chan<- fetchResult) {
	// Check if response is ok
	if !response.Get("ok").Bool() {
		status := response.Get("status").Int()
		statusText := response.Get("statusText").String()
		resultChan <- fetchResult{error: fmt.Errorf("HTTP %d: %s", status, statusText)}
		return
	}

	// Get response text
	textPromise := response.Call("json")

	// Handle text promise with proper function lifecycle
	var thenFunc js.Func
	thenFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer thenFunc.Release()
		if len(args) > 0 {
			jsonResponse := args[0]

			// Convert JS object to Go map
			result := jsValueToInterface(jsonResponse)

			// Check if the response indicates success
			if respMap, ok := result.(map[string]interface{}); ok {
				if success, exists := respMap["success"]; exists && success == true {
					if resultData, exists := respMap["result"]; exists {
						resultChan <- fetchResult{data: resultData}
						return nil
					}
				}
				if errorMsg, exists := respMap["error"]; exists {
					resultChan <- fetchResult{error: fmt.Errorf("server error: %v", errorMsg)}
					return nil
				}
			}

			resultChan <- fetchResult{data: result}
		}
		return nil
	})

	var catchFunc js.Func
	catchFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer catchFunc.Release()
		if len(args) > 0 {
			err := fmt.Errorf("response parsing error: %s", args[0].String())
			resultChan <- fetchResult{error: err}
		}
		return nil
	})

	textPromise.Call("then", thenFunc).Call("catch", catchFunc)
}

// jsValueToInterface converts a JavaScript value to a Go interface{}
func jsValueToInterface(val js.Value) interface{} {
	switch val.Type() {
	case js.TypeString:
		return val.String()
	case js.TypeNumber:
		return val.Float()
	case js.TypeBoolean:
		return val.Bool()
	case js.TypeObject:
		if val.IsNull() {
			return nil
		}
		// Handle arrays
		if val.Get("length").Type() != js.TypeUndefined {
			length := val.Get("length").Int()
			arr := make([]interface{}, length)
			for i := 0; i < length; i++ {
				arr[i] = jsValueToInterface(val.Index(i))
			}
			return arr
		}
		// Handle objects
		obj := make(map[string]interface{})
		keys := js.Global().Get("Object").Call("keys", val)
		for i := 0; i < keys.Get("length").Int(); i++ {
			key := keys.Index(i).String()
			obj[key] = jsValueToInterface(val.Get(key))
		}
		return obj
	case js.TypeNull, js.TypeUndefined:
		return nil
	default:
		return val.String()
	}
}

// Convenience functions for common patterns

// CallString calls a function and expects a string result
func (c *Client) CallString(ctx context.Context, serviceName, functionName string, args ...interface{}) (string, error) {
	result, err := c.Call(ctx, serviceName, functionName, args...)
	if err != nil {
		return "", err
	}
	if str, ok := result.(string); ok {
		return str, nil
	}
	return fmt.Sprintf("%v", result), nil
}

// CallMap calls a function and expects a map result
func (c *Client) CallMap(ctx context.Context, serviceName, functionName string, args ...interface{}) (map[string]interface{}, error) {
	result, err := c.Call(ctx, serviceName, functionName, args...)
	if err != nil {
		return nil, err
	}
	if m, ok := result.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, fmt.Errorf("result is not a map: %T", result)
}

// CallInt calls a function and expects an integer result
func (c *Client) CallInt(ctx context.Context, serviceName, functionName string, args ...interface{}) (int, error) {
	result, err := c.Call(ctx, serviceName, functionName, args...)
	if err != nil {
		return 0, err
	}
	switch v := result.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		// Try to parse as number if it's a string
		return 0, fmt.Errorf("cannot convert string to int: %s", v)
	default:
		return 0, fmt.Errorf("result is not a number: %T", result)
	}
}

// Global client instance for convenience
var defaultClient *Client

// SetDefaultClient sets the global default client
func SetDefaultClient(baseURL string) {
	defaultClient = NewClient(baseURL)
}

// GetDefaultClient returns the default client
func GetDefaultClient() *Client {
	return defaultClient
}

// Convenience functions using the default client

// Call is a convenience function for calling server functions with the default client
func Call(ctx context.Context, serviceName, functionName string, args ...interface{}) (interface{}, error) {
	if defaultClient == nil {
		// Auto-initialize with current origin if not configured
		fmt.Printf("🔗 Auto-initializing gRPC client with empty baseURL\n")
		defaultClient = NewClient("")
		fmt.Printf("🔗 Golem gRPC client auto-initialized (baseURL: '%s', timeout: %v)\n", defaultClient.baseURL, defaultClient.timeout)
	}
	return defaultClient.Call(ctx, serviceName, functionName, args...)
}

// CallString is a convenience function for calling server functions that return strings
func CallString(ctx context.Context, serviceName, functionName string, args ...interface{}) (string, error) {
	if defaultClient == nil {
		// Auto-initialize with current origin if not configured
		fmt.Printf("🔗 Auto-initializing gRPC client with empty baseURL\n")
		defaultClient = NewClient("")
		fmt.Printf("🔗 Golem gRPC client auto-initialized (baseURL: '%s', timeout: %v)\n", defaultClient.baseURL, defaultClient.timeout)
	}
	return defaultClient.CallString(ctx, serviceName, functionName, args...)
}

// CallMap is a convenience function for calling server functions that return maps
func CallMap(ctx context.Context, serviceName, functionName string, args ...interface{}) (map[string]interface{}, error) {
	if defaultClient == nil {
		// Auto-initialize with current origin if not configured
		fmt.Printf("🔗 Auto-initializing gRPC client with empty baseURL\n")
		defaultClient = NewClient("")
		fmt.Printf("🔗 Golem gRPC client auto-initialized (baseURL: '%s', timeout: %v)\n", defaultClient.baseURL, defaultClient.timeout)
	}
	return defaultClient.CallMap(ctx, serviceName, functionName, args...)
}

// CallInt is a convenience function for calling server functions that return integers
func CallInt(ctx context.Context, serviceName, functionName string, args ...interface{}) (int, error) {
	if defaultClient == nil {
		// Auto-initialize with current origin if not configured
		fmt.Printf("🔗 Auto-initializing gRPC client with empty baseURL\n")
		defaultClient = NewClient("")
		fmt.Printf("🔗 Golem gRPC client auto-initialized (baseURL: '%s', timeout: %v)\n", defaultClient.baseURL, defaultClient.timeout)
	}
	return defaultClient.CallInt(ctx, serviceName, functionName, args...)
}
