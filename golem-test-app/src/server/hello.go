package server

import (
	"context"
	"fmt"

	"github.com/Nu11ified/golem/functions"
)

// Register all server functions during package initialization
func init() {
	// Register each function with the global registry
	functions.Register("server", "Hello", Hello)
	functions.Register("server", "GetUserProfile", GetUserProfile)
	functions.Register("server", "Calculate", Calculate)
}

// Hello is a server function that can be called from the client
func Hello(name string) string {
	return fmt.Sprintf("Hello, %s! This message is from the Go server.", name)
}

// GetUserProfile fetches user profile data
func GetUserProfile(ctx context.Context, userID int) (map[string]interface{}, error) {
	// Simulate database lookup
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID: %d", userID)
	}

	return map[string]interface{}{
		"id":    userID,
		"name":  "John Doe",
		"email": "john@example.com",
		"role":  "admin",
	}, nil
}

// Calculate performs basic math operations (for demo purposes)
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

type UserProfile struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
