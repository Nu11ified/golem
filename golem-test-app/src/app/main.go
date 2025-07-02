package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/state"
)

// Post represents the structure of a post from the API
type Post struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

func App() *dom.Element {
	// --- STATE ---
	// Create observables for our application's state
	posts := state.NewObservable([]Post{})
	isLoading := state.NewObservable(false)
	errorMessage := state.NewObservable("")

	// --- UI ELEMENTS ---
	// Create a container for the list of posts
	postListElement := dom.Ul()

	// Create a loading indicator
	loadingIndicator := dom.P("Loading...")

	// Create an error message container
	errorElement := dom.P()

	// --- SUBSCRIPTIONS (Reactivity) ---
	// Subscribe to the 'posts' state to update the list when data arrives
	posts.Subscribe(func(newPosts, oldPosts []Post) {
		var newChildren []*dom.Element
		for _, post := range newPosts {
			li := dom.Li(fmt.Sprintf("#%d: %s", post.ID, post.Title))
			newChildren = append(newChildren, li)
		}
		// Replace the children slice and call Render() to update the DOM
		postListElement.Children = newChildren
		postListElement.Render()
	})

	// Subscribe to 'isLoading' to show/hide the loading indicator
	isLoading.Subscribe(func(loading, _ bool) {
		if loading {
			loadingIndicator.Update(map[string]interface{}{"style": "display: block;"})
		} else {
			loadingIndicator.Update(map[string]interface{}{"style": "display: none;"})
		}
	})

	// Subscribe to 'errorMessage' to show/hide the error message
	errorMessage.Subscribe(func(newError, _ string) {
		if newError != "" {
			errorElement.Update(map[string]interface{}{
				"textContent": newError,
				"style":       "color: red;",
			})
		} else {
			errorElement.Update(map[string]interface{}{
				"textContent": "",
				"style":       "display: none;",
			})
		}
	})

	// --- RENDER ---
	return dom.Div(
		dom.Class("app"),
		dom.H1("Golem Post Fetcher 🚀"),
		dom.P("A real-world example of fetching data from an API."),

		// The button that triggers the data fetch
		dom.Button(
			"Fetch Posts",
			dom.OnClick(func() {
				// Set loading state to true and clear any old errors
				isLoading.Set(true)
				errorMessage.Set("")

				// Run the network request in a goroutine to avoid blocking the UI
				go func() {
					resp, err := http.Get("https://jsonplaceholder.typicode.com/posts")
					if err != nil {
						errorMessage.Set(fmt.Sprintf("Failed to fetch posts: %v", err))
						isLoading.Set(false)
						return
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						errorMessage.Set(fmt.Sprintf("API request failed with status: %s", resp.Status))
						isLoading.Set(false)
						return
					}

					body, err := io.ReadAll(resp.Body)
					if err != nil {
						errorMessage.Set(fmt.Sprintf("Failed to read response body: %v", err))
						isLoading.Set(false)
						return
					}

					var fetchedPosts []Post
					if err := json.Unmarshal(body, &fetchedPosts); err != nil {
						errorMessage.Set(fmt.Sprintf("Failed to parse posts JSON: %v", err))
						isLoading.Set(false)
						return
					}

					// Update the posts observable with the new data
					posts.Set(fetchedPosts)
					isLoading.Set(false)
				}()
			}),
		),

		// Placeholders for our dynamic content
		loadingIndicator,
		errorElement,
		postListElement,
	)
}

func main() {
	dom.Render(App(), "#app")
	select {}
}
