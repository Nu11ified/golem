//go:build js && wasm

package components

import (
	"context"
	"fmt"
	"time"

	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/grpc"
	"github.com/Nu11ified/golem/state"
)

type ServerDemoState struct {
	Name          string
	HelloResponse string
	UserID        string
	UserProfile   map[string]interface{}
	IsLoading     bool
	ErrorMessage  string
	LastCallTime  string
}

func ServerDemoComponent() *dom.Element {
	// Initialize component state
	stateManager := state.NewReactiveState(&ServerDemoState{
		Name:         "World",
		UserID:       "123",
		IsLoading:    false,
		ErrorMessage: "",
	})

	// Helper function to update state safely
	updateState := func(updater func(*ServerDemoState)) {
		stateManager.Update(func(current interface{}) interface{} {
			state := current.(*ServerDemoState)
			updater(state)
			return state
		})
	}

	// Function to call the Hello server function
	callHelloFunction := func() {
		currentState := stateManager.Get().(*ServerDemoState)
		fmt.Printf("🔄 Calling Hello function with name: %s\n", currentState.Name)

		updateState(func(s *ServerDemoState) {
			s.IsLoading = true
			s.ErrorMessage = ""
		})

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			fmt.Printf("📡 Making gRPC call to Hello function...\n")
			result, err := grpc.CallString(ctx, "server", "Hello", currentState.Name)

			updateState(func(s *ServerDemoState) {
				s.IsLoading = false
				s.LastCallTime = time.Now().Format("15:04:05")
				if err != nil {
					fmt.Printf("❌ Error calling Hello: %v\n", err)
					s.ErrorMessage = fmt.Sprintf("Error calling Hello: %v", err)
					s.HelloResponse = ""
				} else {
					fmt.Printf("✅ Hello response: %s\n", result)
					s.HelloResponse = result
					s.ErrorMessage = ""
				}
			})
		}()
	}

	// Function to call the GetUserProfile server function
	callUserProfileFunction := func() {
		currentState := stateManager.Get().(*ServerDemoState)

		updateState(func(s *ServerDemoState) {
			s.IsLoading = true
			s.ErrorMessage = ""
		})

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// Convert userID string to int
			userID := 123 // Default value
			if currentState.UserID != "" {
				fmt.Sscanf(currentState.UserID, "%d", &userID)
			}

			fmt.Printf("📡 Making gRPC call to GetUserProfile with userID: %d\n", userID)
			result, err := grpc.CallMap(ctx, "server", "GetUserProfile", userID)

			updateState(func(s *ServerDemoState) {
				s.IsLoading = false
				s.LastCallTime = time.Now().Format("15:04:05")
				if err != nil {
					fmt.Printf("❌ Error calling GetUserProfile: %v\n", err)
					s.ErrorMessage = fmt.Sprintf("Error calling GetUserProfile: %v", err)
					s.UserProfile = nil
				} else {
					fmt.Printf("✅ GetUserProfile response: %+v\n", result)
					s.UserProfile = result
					s.ErrorMessage = ""
				}
			})
		}()
	}

	// Function to test server connectivity
	testServerConnection := func() {
		fmt.Printf("🔄 Testing server connection...\n")

		updateState(func(s *ServerDemoState) {
			s.IsLoading = true
			s.ErrorMessage = ""
		})

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Try to call a simple function to test connectivity
			fmt.Printf("📡 Making test gRPC call to Hello function...\n")
			_, err := grpc.CallString(ctx, "server", "Hello", "Connection Test")

			updateState(func(s *ServerDemoState) {
				s.IsLoading = false
				s.LastCallTime = time.Now().Format("15:04:05")
				if err != nil {
					fmt.Printf("❌ Connection test failed: %v\n", err)
					s.ErrorMessage = fmt.Sprintf("❌ Server not reachable: %v", err)
				} else {
					fmt.Printf("✅ Connection test successful!\n")
					s.ErrorMessage = "✅ Server connection successful!"
				}
			})
		}()
	}

	return stateManager.WithState(func(s interface{}) *dom.Element {
		state := s.(*ServerDemoState)

		// Build the user profile display
		var userProfileDisplay *dom.Element
		if state.UserProfile != nil {
			userProfileDisplay = dom.Div(
				dom.Class("user-profile"),
				dom.H4(dom.Text("👤 User Profile:")),
				dom.P(dom.Text(fmt.Sprintf("ID: %v", state.UserProfile["id"]))),
				dom.P(dom.Text(fmt.Sprintf("Name: %v", state.UserProfile["name"]))),
				dom.P(dom.Text(fmt.Sprintf("Email: %v", state.UserProfile["email"]))),
			)
		} else {
			userProfileDisplay = dom.Div()
		}

		// Build the hello response display
		var helloDisplay *dom.Element
		if state.HelloResponse != "" {
			helloDisplay = dom.Div(
				dom.Class("hello-response"),
				dom.H4(dom.Text("💬 Server Response:")),
				dom.P(dom.Text(state.HelloResponse)),
			)
		} else {
			helloDisplay = dom.Div()
		}

		// Build status display
		var statusDisplay *dom.Element
		if state.ErrorMessage != "" {
			statusDisplay = dom.Div(
				dom.Class("status-message"),
				dom.P(dom.Text(state.ErrorMessage)),
			)
		} else if state.LastCallTime != "" {
			statusDisplay = dom.Div(
				dom.Class("status-message success"),
				dom.P(dom.Text(fmt.Sprintf("✅ Last call successful at %s", state.LastCallTime))),
			)
		} else {
			statusDisplay = dom.Div()
		}

		// Loading indicator
		var loadingIndicator *dom.Element
		if state.IsLoading {
			loadingIndicator = dom.Div(
				dom.Class("loading"),
				dom.P(dom.Text("🔄 Calling server function...")),
			)
		} else {
			loadingIndicator = dom.Div()
		}

		return dom.Div(
			dom.Class("server-demo-app"),
			dom.H2(dom.Text("🚀 Server Function Demo")),
			dom.P(dom.Text("This demo showcases calling Go server functions from the frontend using gRPC.")),

			// Connection test section
			dom.Div(
				dom.Class("demo-section"),
				dom.H3(dom.Text("🔌 Connection Test")),
				dom.Button(
					dom.Text("Test Server Connection"),
					dom.OnClick(func() { testServerConnection() }),
					dom.If(state.IsLoading, dom.Disabled(true)),
				),
			),

			// Hello function demo
			dom.Div(
				dom.Class("demo-section"),
				dom.H3(dom.Text("👋 Hello Function")),
				dom.Div(
					dom.Class("input-group"),
					dom.Label(dom.Text("Your Name:")),
					dom.Input(
						dom.Type("text"),
						dom.Value(state.Name),
						dom.Placeholder("Enter your name"),
						dom.OnInput(func(value string) {
							updateState(func(s *ServerDemoState) {
								s.Name = value
							})
						}),
					),
					dom.Button(
						dom.Text("Call Hello Function"),
						dom.OnClick(func() { callHelloFunction() }),
						dom.If(state.IsLoading, dom.Disabled(true)),
					),
				),
				helloDisplay,
			),

			// User Profile function demo
			dom.Div(
				dom.Class("demo-section"),
				dom.H3(dom.Text("👤 User Profile Function")),
				dom.Div(
					dom.Class("input-group"),
					dom.Label(dom.Text("User ID:")),
					dom.Input(
						dom.Type("number"),
						dom.Value(state.UserID),
						dom.Placeholder("Enter user ID"),
						dom.OnInput(func(value string) {
							updateState(func(s *ServerDemoState) {
								s.UserID = value
							})
						}),
					),
					dom.Button(
						dom.Text("Get User Profile"),
						dom.OnClick(func() { callUserProfileFunction() }),
						dom.If(state.IsLoading, dom.Disabled(true)),
					),
				),
				userProfileDisplay,
			),

			// Status and loading
			loadingIndicator,
			statusDisplay,

			// Information section
			dom.Div(
				dom.Class("info-section"),
				dom.H3(dom.Text("ℹ️ How It Works")),
				dom.Ul(
					dom.Li(dom.Text("Frontend Go/WASM code calls server functions seamlessly")),
					dom.Li(dom.Text("Arguments are automatically marshaled to JSON")),
					dom.Li(dom.Text("Server functions execute and return results")),
					dom.Li(dom.Text("Results are unmarshaled back to Go types")),
					dom.Li(dom.Text("Error handling works across the network boundary")),
				),
			),
		)
	})
}
