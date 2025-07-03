# gRPC Function Calling Guide

This guide explains how to use Golem's seamless server function calling system, which enables you to call Go server functions directly from your frontend Go/WASM code as if they were local functions.

## Overview

The Golem framework provides a powerful gRPC-based system that allows you to:
- Call server functions from frontend code seamlessly
- Automatic argument marshaling and result unmarshaling
- Type-safe function calls with proper error handling
- Support for complex data types (strings, numbers, objects, arrays)
- Automatic function discovery and registration

## Architecture

```
Frontend (Go/WASM) → HTTP/JSON → gRPC Function Handler → Go Server Functions
```

The system works by:
1. **Function Registration**: Server functions are automatically registered in a function registry
2. **Frontend Client**: A gRPC client in the frontend makes HTTP requests to call server functions
3. **HTTP Bridge**: An HTTP handler converts JSON requests to gRPC calls
4. **Function Execution**: The registry executes the requested function and returns the result

## Quick Start

### 1. Define Server Functions

Create server functions in your Go server code:

```go
// server/api.go
package server

func Hello(name string) string {
    return fmt.Sprintf("Hello, %s! This message is from the Go server.", name)
}

func GetUserProfile(userID int) (map[string]interface{}, error) {
    // Simulate database lookup
    if userID <= 0 {
        return nil, fmt.Errorf("invalid user ID")
    }
    
    return map[string]interface{}{
        "id":    userID,
        "name":  "John Doe",
        "email": "john@example.com",
    }, nil
}

func Calculate(a, b float64, operation string) (float64, error) {
    switch operation {
    case "add":
        return a + b, nil
    case "subtract":
        return a - b, nil
    case "multiply":
        return a * b, nil
    case "divide":
        if b == 0 {
            return 0, fmt.Errorf("division by zero")
        }
        return a / b, nil
    default:
        return 0, fmt.Errorf("unknown operation: %s", operation)
    }
}
```

### 2. Register Functions in Server

```go
// main.go or server setup
func main() {
    config := &config.Config{
        // ... your config
    }
    
    server := server.NewServer(config)
    
    // Functions are automatically registered through the registry
    server.Start() // This will start both HTTP and gRPC servers
}
```

### 3. Call Functions from Frontend

```go
//go:build js && wasm

package main

import (
    "context"
    "fmt"
    "log"
    "syscall/js"
    "time"

    "github.com/Nu11ified/golem/grpc"
)

func main() {
    // Initialize the default client
    grpc.SetDefaultClient("http://localhost:3000")

    // Call server functions seamlessly
    ctx := context.Background()
    
    // Example 1: Simple string function
    result, err := grpc.CallString(ctx, "server", "Hello", "World")
    if err != nil {
        log.Printf("Error: %v", err)
        return
    }
    fmt.Printf("Server says: %s\n", result)
    
    // Example 2: Function returning complex data
    profile, err := grpc.CallMap(ctx, "server", "GetUserProfile", 123)
    if err != nil {
        log.Printf("Error: %v", err)
        return
    }
    fmt.Printf("User: %s (%s)\n", profile["name"], profile["email"])
    
    // Example 3: Function with multiple parameters
    calcResult, err := grpc.Call(ctx, "server", "Calculate", 10.0, 5.0, "add")
    if err != nil {
        log.Printf("Error: %v", err)
        return
    }
    fmt.Printf("10 + 5 = %v\n", calcResult)
}
```

## Function Registration

Functions are registered automatically when the server starts. You can also register functions manually:

```go
registry := functions.NewRegistry()

// Register a function manually
err := registry.RegisterFunction("math", "Add", func(a, b int) int {
    return a + b
})
```

### Function Requirements

Server functions must follow these patterns:

```go
// Pattern 1: Simple function
func FunctionName(args...) returnType

// Pattern 2: Function with error handling
func FunctionName(args...) (returnType, error)

// Pattern 3: Function with context (context is automatically provided)
func FunctionName(ctx context.Context, args...) (returnType, error)
```

**Supported Types:**
- Primitive types: `string`, `int`, `float64`, `bool`
- Complex types: `map[string]interface{}`, `[]interface{}`
- Custom structs (must be JSON serializable)
- Error handling through `(result, error)` return pattern

## Frontend Client API

### Initialize Client

```go
// Set default client (used by convenience functions)
grpc.SetDefaultClient("http://localhost:3000")

// Or create custom client
client := grpc.NewClient("http://localhost:8080")
client.SetTimeout(5 * time.Second)
```

### Call Functions

```go
// Generic call (returns interface{})
result, err := grpc.Call(ctx, "serviceName", "functionName", arg1, arg2, ...)

// Type-specific convenience functions
str, err := grpc.CallString(ctx, "server", "Hello", "World")
num, err := grpc.CallInt(ctx, "math", "Add", 5, 3)
data, err := grpc.CallMap(ctx, "server", "GetUserProfile", 123)

// Custom client
result, err := client.Call(ctx, "server", "Hello", "World")
```

### Error Handling

```go
result, err := grpc.CallString(ctx, "server", "Hello", "World")
if err != nil {
    // Handle different types of errors
    switch {
    case strings.Contains(err.Error(), "timeout"):
        log.Println("Request timed out")
    case strings.Contains(err.Error(), "not found"):
        log.Println("Function not found")
    default:
        log.Printf("Function error: %v", err)
    }
    return
}
```

## JavaScript Integration

Expose server functions to JavaScript:

```go
//go:build js && wasm

func main() {
    grpc.SetDefaultClient("http://localhost:3000")
    
    // Expose functions to JavaScript
    js.Global().Set("callServerHello", js.FuncOf(callServerHello))
    js.Global().Set("getUserProfile", js.FuncOf(getUserProfile))
    
    // Keep the program running
    select {}
}

func callServerHello(this js.Value, p []js.Value) interface{} {
    if len(p) == 0 {
        return js.ValueOf("Error: No name provided")
    }
    
    name := p[0].String()
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    result, err := grpc.CallString(ctx, "server", "Hello", name)
    if err != nil {
        return js.ValueOf(fmt.Sprintf("Error: %v", err))
    }
    
    return js.ValueOf(result)
}
```

Then in JavaScript:

```javascript
// Call server functions from browser JavaScript
const greeting = await callServerHello("World");
console.log(greeting); // "Hello, World! This message is from the Go server."

const user = await getUserProfile(123);
console.log(user.name); // "John Doe"
```

## Configuration

### Server Configuration

```go
config := &config.Config{
    Server: config.ServerConfig{
        GRPC: config.GRPCConfig{
            Port: 50051, // gRPC port
        },
        Functions: "src/server", // Directory to scan for functions
    },
    Output: "./dist", // Static files directory
}
```

### Development vs Production

**Development Mode:**
- Hot reload enabled
- Function registry updates automatically
- Debug logging enabled

**Production Mode:**
- Optimized performance
- Function registry cached
- Minimal logging

## Advanced Usage

### Custom Function Services

Group functions into logical services:

```go
// Register functions under different services
registry.RegisterFunction("auth", "Login", LoginFunc)
registry.RegisterFunction("auth", "Logout", LogoutFunc)
registry.RegisterFunction("data", "GetUsers", GetUsersFunc)
registry.RegisterFunction("data", "CreateUser", CreateUserFunc)
```

```go
// Call functions by service
user, err := grpc.CallMap(ctx, "auth", "Login", username, password)
users, err := grpc.CallMap(ctx, "data", "GetUsers")
```

### Batch Operations

```go
// Call multiple functions in sequence
ctx := context.Background()

results := make([]interface{}, 0)
for i := 1; i <= 5; i++ {
    profile, err := grpc.CallMap(ctx, "server", "GetUserProfile", i*100)
    if err != nil {
        log.Printf("Failed to get profile %d: %v", i*100, err)
        continue
    }
    results = append(results, profile)
}

fmt.Printf("Retrieved %d profiles\n", len(results))
```

### Type-Safe Wrappers

Create type-safe wrappers for your functions:

```go
type UserProfile struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func GetUserProfileTyped(ctx context.Context, userID int) (*UserProfile, error) {
    result, err := grpc.CallMap(ctx, "server", "GetUserProfile", userID)
    if err != nil {
        return nil, err
    }
    
    // Convert map to struct
    jsonData, err := json.Marshal(result)
    if err != nil {
        return nil, err
    }
    
    var profile UserProfile
    if err := json.Unmarshal(jsonData, &profile); err != nil {
        return nil, err
    }
    
    return &profile, nil
}
```

## Performance

The function calling system is highly optimized:

- **~47 microseconds** per function call (including full HTTP roundtrip)
- **~13KB memory** per call with 164 allocations
- Supports **high concurrency** with minimal overhead
- **JSON serialization** for maximum compatibility

### Performance Tips

1. **Reuse clients**: Create one client and reuse it
2. **Use contexts**: Set appropriate timeouts
3. **Batch operations**: Group multiple calls when possible
4. **Type-specific calls**: Use `CallString`, `CallInt`, etc. for better performance

## Testing

Test your server functions:

```go
func TestServerFunctions(t *testing.T) {
    registry := functions.NewRegistry()
    registry.RegisterFunction("server", "Hello", Hello)
    
    grpcServer := functions.NewGRPCServer(registry)
    testServer := httptest.NewServer(http.HandlerFunc(grpcServer.HTTPHandler()))
    defer testServer.Close()
    
    // Test function call
    requestData := map[string]interface{}{
        "functionName": "Hello",
        "serviceName":  "server",
        "args":         []interface{}{"Test"},
    }
    
    jsonData, _ := json.Marshal(requestData)
    resp, err := http.Post(testServer.URL, "application/json", strings.NewReader(string(jsonData)))
    assert.NoError(t, err)
    
    var response struct {
        Success bool        `json:"success"`
        Result  interface{} `json:"result"`
    }
    json.NewDecoder(resp.Body).Decode(&response)
    
    assert.True(t, response.Success)
    assert.Equal(t, "Hello, Test! This message is from the Go server.", response.Result)
}
```

## Security Considerations

1. **Authentication**: Add authentication middleware to protect function endpoints
2. **Authorization**: Implement role-based access control for functions
3. **Rate Limiting**: Add rate limiting to prevent abuse
4. **Input Validation**: Validate all function inputs
5. **CORS**: Configure CORS properly for production

```go
// Example authentication middleware
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if !validateToken(token) {
            w.WriteHeader(http.StatusUnauthorized)
            return
        }
        next(w, r)
    }
}

// Apply to function endpoint
mux.HandleFunc("/api/functions", authMiddleware(grpcServer.HTTPHandler()))
```

## Troubleshooting

### Common Issues

1. **Function not found**
   - Ensure function is exported (starts with capital letter)
   - Check function registration
   - Verify service name matches

2. **Serialization errors**
   - Ensure all types are JSON serializable
   - Check for circular references in complex types
   - Use simple types when possible

3. **Timeout errors**
   - Increase client timeout
   - Check server function performance
   - Verify network connectivity

4. **Type conversion errors**
   - Use correct argument types
   - Check function signature matches call

### Debug Mode

Enable debug logging:

```go
log.SetLevel(log.DebugLevel)
```

This will show detailed information about function calls, argument conversion, and errors.

## Examples

See the `examples/` directory for complete working examples:
- `examples/frontend_usage.go` - Frontend client usage
- `test/grpc_integration_test.go` - Comprehensive testing examples

## Next Steps

- Explore the example applications
- Try creating your own server functions
- Test the system with your existing Go backend
- Consider adding authentication and authorization
- Scale to production with proper monitoring

The gRPC function calling system makes it incredibly easy to build full-stack Go applications with seamless communication between frontend and backend code. 