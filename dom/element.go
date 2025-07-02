//go:build js && wasm

package dom

import (
	"fmt"
	"reflect"
	"syscall/js"
)

// Element represents a virtual DOM element
type Element struct {
	Type          string
	Props         map[string]interface{}
	Children      []*Element
	EventHandlers map[string]func()
	JSElement     js.Value
}

// Attribute represents an HTML attribute
type Attribute struct {
	Name  string
	Value interface{}
}

// NewElement creates a new virtual DOM element with mixed arguments
func NewElement(tagType string, args ...interface{}) *Element {
	props := make(map[string]interface{})
	eventHandlers := make(map[string]func())
	children := make([]*Element, 0)

	for _, arg := range args {
		switch v := arg.(type) {
		case Attribute:
			if v.Name == "onclick" {
				if handler, ok := v.Value.(func()); ok {
					eventHandlers["click"] = handler
				}
			} else if v.Name != "" { // Skip empty attributes from If() function
				props[v.Name] = v.Value
			}
		case *Element:
			children = append(children, v)
		case string:
			// Text content
			textElement := &Element{
				Type:          "text",
				Props:         map[string]interface{}{"textContent": v},
				Children:      make([]*Element, 0),
				EventHandlers: make(map[string]func()),
			}
			children = append(children, textElement)
		}
	}

	return &Element{
		Type:          tagType,
		Props:         props,
		Children:      children,
		EventHandlers: eventHandlers,
	}
}

// AddChild adds a child element
func (e *Element) AddChild(child *Element) {
	e.Children = append(e.Children, child)
}

// Render creates or updates the DOM element
func (e *Element) Render() js.Value {
	// Handle text nodes
	if e.Type == "text" {
		if e.JSElement.IsUndefined() {
			doc := js.Global().Get("document")
			textContent := fmt.Sprintf("%v", e.Props["textContent"])
			e.JSElement = doc.Call("createTextNode", textContent)
		}
		return e.JSElement
	}

	// Create DOM element if it doesn't exist
	if e.JSElement.IsUndefined() {
		doc := js.Global().Get("document")
		e.JSElement = doc.Call("createElement", e.Type)

		// Set properties
		for name, value := range e.Props {
			switch name {
			case "class":
				e.JSElement.Set("className", value)
			case "id":
				e.JSElement.Set("id", value)
			case "textContent":
				e.JSElement.Set("textContent", value)
			default:
				e.JSElement.Call("setAttribute", name, fmt.Sprintf("%v", value))
			}
		}

		// Add event listeners
		for event, handler := range e.EventHandlers {
			jsHandler := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				handler()
				return nil
			})
			e.JSElement.Call("addEventListener", event, jsHandler)
		}
	}

	// Clear existing children
	e.JSElement.Set("innerHTML", "")

	// Render children
	for _, child := range e.Children {
		childElement := child.Render()
		e.JSElement.Call("appendChild", childElement)
	}

	return e.JSElement
}

// Update updates the element with new props
func (e *Element) Update(newProps map[string]interface{}) {
	// Compare and update only changed properties
	for name, newValue := range newProps {
		if oldValue, exists := e.Props[name]; !exists || !reflect.DeepEqual(oldValue, newValue) {
			e.Props[name] = newValue

			// Update DOM property
			if !e.JSElement.IsUndefined() {
				switch name {
				case "class":
					e.JSElement.Set("className", newValue)
				case "id":
					e.JSElement.Set("id", newValue)
				case "textContent":
					e.JSElement.Set("textContent", newValue)
				default:
					e.JSElement.Call("setAttribute", name, fmt.Sprintf("%v", newValue))
				}
			}
		}
	}
}

// Helpers for creating common attributes
func Class(className string) Attribute {
	return Attribute{Name: "class", Value: className}
}

func Id(id string) Attribute {
	return Attribute{Name: "id", Value: id}
}

func Text(text interface{}) Attribute {
	return Attribute{Name: "textContent", Value: fmt.Sprintf("%v", text)}
}

func OnClick(handler func()) Attribute {
	return Attribute{Name: "onclick", Value: handler}
}

func Disabled(disabled bool) Attribute {
	return Attribute{Name: "disabled", Value: disabled}
}

func If(condition bool, attr Attribute) Attribute {
	if condition {
		return attr
	}
	return Attribute{Name: "", Value: nil}
}

// Common HTML elements
func Div(args ...interface{}) *Element {
	return NewElement("div", args...)
}

func H1(args ...interface{}) *Element {
	return NewElement("h1", args...)
}

func H2(args ...interface{}) *Element {
	return NewElement("h2", args...)
}

func P(args ...interface{}) *Element {
	return NewElement("p", args...)
}

func Button(args ...interface{}) *Element {
	return NewElement("button", args...)
}

func Input(args ...interface{}) *Element {
	return NewElement("input", args...)
}

func Span(args ...interface{}) *Element {
	return NewElement("span", args...)
}

func A(args ...interface{}) *Element {
	return NewElement("a", args...)
}

func Img(args ...interface{}) *Element {
	return NewElement("img", args...)
}

func Ul(args ...interface{}) *Element {
	return NewElement("ul", args...)
}

func Li(args ...interface{}) *Element {
	return NewElement("li", args...)
}

// Render renders an element tree to a target selector
func Render(element *Element, selector string) {
	doc := js.Global().Get("document")
	target := doc.Call("querySelector", selector)

	if target.IsNull() {
		fmt.Printf("Target element not found: %s\n", selector)
		return
	}

	// Clear target
	target.Set("innerHTML", "")

	// Render and append
	renderedElement := element.Render()
	target.Call("appendChild", renderedElement)
}

// Alert shows a browser alert
func Alert(message string) {
	js.Global().Call("alert", message)
}
