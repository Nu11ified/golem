# Server Function Demo

This demo showcases Golem's seamless server function calling capabilities. You can call Go server functions directly from the frontend as if they were local functions!

## What's Included

### Server Functions (`src/server/hello.go`)
- **Hello(name string)**: Simple greeting function
- **GetUserProfile(ctx, userID int)**: Returns user profile data
- **Calculate(a, b float64, operation string)**: Basic math operations

### Demo Component (`src/components/serverdemo.go`)
- Interactive UI for testing server functions
- Real-time error handling and loading states
- Demonstrates different return types (strings, objects, numbers)

## How to Run

1. **Start the development server:**
   ```bash
   cd golem-test-app
   go run ../cmd/golem dev
   ```

2. **Open your browser:**
   Navigate to the URL shown in the terminal (usually `http://localhost:3000`)

3. **Try the demo:**
   - Test server connection
   - Call the Hello function with your name
   - Get user profile data
   - See real-time responses and error handling

## Features Demonstrated

### ✅ Seamless Function Calls
```go
// Frontend code calling server function
result, err := grpc.CallString(ctx, "server", "Hello", "World")
```

### ✅ Automatic Type Marshaling
- Strings, numbers, and complex objects are automatically converted
- Error handling across the network boundary
- Type-safe function calls

### ✅ Real-time UI Updates
- Loading states during function calls
- Error messages displayed in UI
- Successful responses shown immediately

### ✅ Multiple Function Patterns
- Simple functions returning strings
- Functions returning complex objects
- Functions with error handling
- Functions with multiple parameters

## Architecture

```
Frontend Component → gRPC Client → HTTP/JSON → Server Function Registry → Go Functions
```

1. **Frontend**: Go/WASM component calls server functions
2. **gRPC Client**: Handles HTTP requests and JSON marshaling
3. **Function Registry**: Automatically discovers and registers server functions
4. **Server Functions**: Regular Go functions that can be called remotely

## Code Structure

```
src/
├── app/main.go           # Main application with all components
├── components/
│   ├── serverdemo.go     # Server function demo component
│   ├── counter.go        # Simple counter component
│   └── todolist.go       # Todo list component
└── server/
    └── hello.go          # Server functions for the demo
```

## Next Steps

1. **Add Your Own Functions**: Create new server functions in `src/server/`
2. **Extend the Demo**: Add more UI interactions in the demo component
3. **Production Ready**: Add authentication, rate limiting, and error handling for production use

The server function calling system makes it incredibly easy to build full-stack Go applications with seamless frontend-backend communication! 