//go:build js && wasm

package components

import (
	"fmt"

	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/state"
)

// CounterComponent creates a simple counter with increment and decrement buttons.
func CounterComponent() *dom.Element {
	// 1. Define reactive state for the counter
	count := state.NewObservable(0)

	// 2. Create a paragraph element that will contain the count text.
	// We hold onto this element to update it directly.
	pElement := dom.P(
		dom.Class("dom-p"),
		fmt.Sprintf("Current count: %d", count.Get()),
	)

	// 3. Subscribe to state changes
	// This function runs whenever count.Set() is called.
	count.Subscribe(func(newValue, _ int) {
		// Update the text content of the first child (the text node) of the paragraph.
		if len(pElement.Children) > 0 {
			pElement.Children[0].Update(map[string]interface{}{
				"textContent": fmt.Sprintf("Current count: %d", newValue),
			})
		}
	})

	// 4. Return the element tree
	return dom.Div(
		dom.Class("counter-app"),
		dom.H2("Counter"),
		pElement, // Embed the reactive paragraph element
		dom.Button(
			"Increment",
			dom.OnClick(func() {
				// Increment the state, which triggers the subscription
				count.Set(count.Get() + 1)
			}),
		),
		dom.Button(
			"Decrement",
			dom.OnClick(func() {
				count.Set(count.Get() - 1)
			}),
		),
	)
}
