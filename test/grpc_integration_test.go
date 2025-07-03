package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Nu11ified/golem/internal/functions"
)

// Server functions to demonstrate the functionality
func Hello(name string) string {
	return fmt.Sprintf("Hello, %s! This message is from the Go server via gRPC.", name)
}

func GetUserProfile(userID int) map[string]interface{} {
	return map[string]interface{}{
		"id":    userID,
		"name":  "John Doe",
		"email": "john@example.com",
		"role":  "admin",
	}
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

// TestGRPCFunctionCalling demonstrates the seamless server function calling
func TestGRPCFunctionCalling(t *testing.T) {
	// Create a function registry and register test functions
	registry := functions.NewRegistry()

	// Register test functions
	if err := registry.RegisterFunction("server", "Hello", Hello); err != nil {
		t.Fatalf("Failed to register Hello function: %v", err)
	}

	if err := registry.RegisterFunction("server", "GetUserProfile", GetUserProfile); err != nil {
		t.Fatalf("Failed to register GetUserProfile function: %v", err)
	}

	if err := registry.RegisterFunction("math", "Calculate", Calculate); err != nil {
		t.Fatalf("Failed to register Calculate function: %v", err)
	}

	// Create a test HTTP server with the gRPC function handler
	grpcServer := functions.NewGRPCServer(registry)
	mux := http.NewServeMux()

	// Add CORS middleware for testing
	corsHandler := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next(w, r)
		}
	}

	mux.HandleFunc("/api/functions", corsHandler(grpcServer.HTTPHandler()))
	mux.HandleFunc("/api/functions/list", corsHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		functions := registry.ListFunctions("")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"functions": functions,
		})
	}))

	// Start test server
	server := httptest.NewServer(mux)
	defer server.Close()

	t.Logf("Test server running at: %s", server.URL)

	// Test 1: Simple string function call
	t.Run("Hello Function", func(t *testing.T) {
		response := callFunction(t, server.URL, "server", "Hello", []interface{}{"World"})

		if !response.Success {
			t.Fatalf("Function call failed: %s", response.Error)
		}

		result, ok := response.Result.(string)
		if !ok {
			t.Fatalf("Expected string result, got %T", response.Result)
		}

		expected := "Hello, World! This message is from the Go server via gRPC."
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}

		t.Logf("✅ Hello function result: %s", result)
	})

	// Test 2: Function returning complex data
	t.Run("GetUserProfile Function", func(t *testing.T) {
		response := callFunction(t, server.URL, "server", "GetUserProfile", []interface{}{123})

		if !response.Success {
			t.Fatalf("Function call failed: %s", response.Error)
		}

		result, ok := response.Result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map result, got %T", response.Result)
		}

		if result["id"].(float64) != 123 {
			t.Errorf("Expected user ID 123, got %v", result["id"])
		}

		if result["name"] != "John Doe" {
			t.Errorf("Expected name 'John Doe', got %v", result["name"])
		}

		t.Logf("✅ GetUserProfile result: %+v", result)
	})

	// Test 3: Function with multiple parameters and error handling
	t.Run("Calculate Function", func(t *testing.T) {
		// Test successful calculation
		response := callFunction(t, server.URL, "math", "Calculate", []interface{}{10.0, 5.0, "add"})

		if !response.Success {
			t.Fatalf("Function call failed: %s", response.Error)
		}

		result, ok := response.Result.(float64)
		if !ok {
			t.Fatalf("Expected float64 result, got %T", response.Result)
		}

		if result != 15.0 {
			t.Errorf("Expected 15.0, got %f", result)
		}

		t.Logf("✅ Calculate (10 + 5) result: %f", result)

		// Test division by zero error
		response = callFunction(t, server.URL, "math", "Calculate", []interface{}{10.0, 0.0, "divide"})

		if response.Success {
			t.Error("Expected function call to fail due to division by zero")
		}

		if response.Error == "" {
			t.Error("Expected error message for division by zero")
		}

		t.Logf("✅ Division by zero properly handled: %s", response.Error)
	})

	// Test 4: List available functions
	t.Run("List Functions", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/functions/list")
		if err != nil {
			t.Fatalf("Failed to list functions: %v", err)
		}
		defer resp.Body.Close()

		var result struct {
			Functions []map[string]interface{} `json:"functions"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode function list: %v", err)
		}

		if len(result.Functions) != 3 {
			t.Errorf("Expected 3 functions, got %d", len(result.Functions))
		}

		t.Logf("✅ Available functions: %d", len(result.Functions))
		for _, fn := range result.Functions {
			t.Logf("  - %s.%s", fn["serviceName"], fn["name"])
		}
	})

	// Test 5: Non-existent function
	t.Run("Non-existent Function", func(t *testing.T) {
		response := callFunction(t, server.URL, "server", "NonExistent", []interface{}{})

		if response.Success {
			t.Error("Expected function call to fail for non-existent function")
		}

		if response.Error == "" {
			t.Error("Expected error message for non-existent function")
		}

		t.Logf("✅ Non-existent function properly handled: %s", response.Error)
	})
}

// FunctionCallResponse represents the response from a function call
type FunctionCallResponse struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result"`
	Error   string      `json:"error"`
}

// callFunction makes a function call to the test server
func callFunction(t *testing.T, baseURL, serviceName, functionName string, args []interface{}) *FunctionCallResponse {
	requestData := map[string]interface{}{
		"functionName": functionName,
		"serviceName":  serviceName,
		"args":         args,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(baseURL+"/api/functions", "application/json",
		strings.NewReader(string(jsonData)))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	var response FunctionCallResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	return &response
}

// BenchmarkFunctionCall benchmarks the function calling performance
func BenchmarkFunctionCall(b *testing.B) {
	// Setup
	registry := functions.NewRegistry()
	registry.RegisterFunction("server", "Hello", Hello)

	grpcServer := functions.NewGRPCServer(registry)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/functions", grpcServer.HTTPHandler())

	server := httptest.NewServer(mux)
	defer server.Close()

	// Benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		response := callFunctionB(b, server.URL, "server", "Hello", []interface{}{"Benchmark"})
		if !response.Success {
			b.Fatalf("Function call failed: %s", response.Error)
		}
	}
}

// callFunction for benchmark (modified to work with testing.B)
func callFunctionB(b *testing.B, baseURL, serviceName, functionName string, args []interface{}) *FunctionCallResponse {
	requestData := map[string]interface{}{
		"functionName": functionName,
		"serviceName":  serviceName,
		"args":         args,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		b.Fatalf("Failed to marshal request: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(baseURL+"/api/functions", "application/json",
		strings.NewReader(string(jsonData)))
	if err != nil {
		b.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	var response FunctionCallResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		b.Fatalf("Failed to decode response: %v", err)
	}

	return &response
}
