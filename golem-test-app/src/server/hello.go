package server

import (
	"context"
	"fmt"
)

// Hello is a server function that can be called from the client
func Hello(name string) string {
	return fmt.Sprintf("Hello, %s! This message is from the Go server.", name)
}

// GetUserProfile fetches user profile data
func GetUserProfile(ctx context.Context, userID int) (*UserProfile, error) {
	// Simulate database lookup
	return &UserProfile{
		ID:   userID,
		Name: "John Doe",
		Email: "john@example.com",
	}, nil
}

type UserProfile struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}