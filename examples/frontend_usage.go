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
	fmt.Println("🚀 Frontend application starting...")

	// Initialize the default client
	grpc.SetDefaultClient("http://localhost:3000")

	// Register JavaScript functions that can be called from the browser
	js.Global().Set("callServerFunction", js.FuncOf(callServerFunction))
	js.Global().Set("callGetUserProfile", js.FuncOf(callGetUserProfile))
	js.Global().Set("callCalculate", js.FuncOf(callCalculate))

	fmt.Println("✅ Frontend application ready!")
	fmt.Println("📡 Server function client configured")
	fmt.Println("🌐 JavaScript functions exposed:")
	fmt.Println("  - window.callServerFunction(name)")
	fmt.Println("  - window.callGetUserProfile(userID)")
	fmt.Println("  - window.callCalculate(a, b, operation)")

	// Keep the Go program running
	select {}
}

// callServerFunction demonstrates calling a simple server function
func callServerFunction(this js.Value, p []js.Value) interface{} {
	// Extract name parameter from JavaScript
	if len(p) == 0 {
		log.Println("❌ No name provided")
		return js.ValueOf("Error: No name provided")
	}

	name := p[0].String()
	log.Printf("📞 Calling server Hello function with name: %s", name)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Call the server function seamlessly
	result, err := grpc.CallString(ctx, "server", "Hello", name)
	if err != nil {
		errorMsg := fmt.Sprintf("Error calling server function: %v", err)
		log.Println("❌", errorMsg)
		return js.ValueOf(errorMsg)
	}

	log.Printf("✅ Server response: %s", result)
	return js.ValueOf(result)
}

// callGetUserProfile demonstrates calling a function that returns complex data
func callGetUserProfile(this js.Value, p []js.Value) interface{} {
	if len(p) == 0 {
		log.Println("❌ No user ID provided")
		return js.ValueOf(map[string]interface{}{"error": "No user ID provided"})
	}

	userID := p[0].Int()
	log.Printf("📞 Calling server GetUserProfile function with ID: %d", userID)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Call the server function and get a map result
	result, err := grpc.CallMap(ctx, "server", "GetUserProfile", userID)
	if err != nil {
		errorMsg := fmt.Sprintf("Error calling server function: %v", err)
		log.Println("❌", errorMsg)
		return js.ValueOf(map[string]interface{}{"error": errorMsg})
	}

	log.Printf("✅ User profile: %+v", result)

	// Convert Go map to JavaScript object
	jsResult := js.Global().Get("Object").New()
	for key, value := range result {
		jsResult.Set(key, value)
	}

	return jsResult
}

// callCalculate demonstrates calling a function with multiple parameters and error handling
func callCalculate(this js.Value, p []js.Value) interface{} {
	if len(p) < 3 {
		log.Println("❌ Insufficient parameters for calculation")
		return js.ValueOf(map[string]interface{}{"error": "Need 3 parameters: a, b, operation"})
	}

	a := p[0].Float()
	b := p[1].Float()
	operation := p[2].String()

	log.Printf("📞 Calling server Calculate function: %.2f %s %.2f", a, operation, b)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Call the server function using the generic Call method
	result, err := grpc.Call(ctx, "math", "Calculate", a, b, operation)
	if err != nil {
		errorMsg := fmt.Sprintf("Error calling server function: %v", err)
		log.Println("❌", errorMsg)
		return js.ValueOf(map[string]interface{}{"error": errorMsg})
	}

	// Convert result to float64
	if floatResult, ok := result.(float64); ok {
		log.Printf("✅ Calculation result: %.2f", floatResult)
		return js.ValueOf(floatResult)
	}

	log.Printf("✅ Calculation result: %v", result)
	return js.ValueOf(result)
}

// Example of a more complex usage pattern
func demonstrateAdvancedUsage() {
	log.Println("🧪 Demonstrating advanced usage patterns...")

	ctx := context.Background()

	// Example 1: Error handling
	_, err := grpc.CallString(ctx, "server", "NonExistentFunction", "test")
	if err != nil {
		log.Printf("✅ Properly handled non-existent function: %v", err)
	}

	// Example 2: Custom client with different timeout
	customClient := grpc.NewClient("http://localhost:8080")
	customClient.SetTimeout(5 * time.Second)

	result, err := customClient.CallString(ctx, "server", "Hello", "Custom Client")
	if err != nil {
		log.Printf("Custom client call failed: %v", err)
	} else {
		log.Printf("Custom client result: %s", result)
	}

	// Example 3: Batch operations (calling multiple functions)
	log.Println("📦 Performing batch operations...")

	results := make([]interface{}, 0)

	// Call multiple functions asynchronously (in a real app, you'd use goroutines)
	for i := 1; i <= 3; i++ {
		profile, err := grpc.CallMap(ctx, "server", "GetUserProfile", i*100)
		if err != nil {
			log.Printf("Failed to get profile for user %d: %v", i*100, err)
			continue
		}
		results = append(results, profile)
	}

	log.Printf("✅ Retrieved %d user profiles", len(results))
}
