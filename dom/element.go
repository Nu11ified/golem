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
	EventHandlers map[string]js.Func
	JSElement     js.Value
}

// Attribute represents an HTML attribute
type Attribute struct {
	Name  string
	Value interface{}
}

// EventAttribute represents an event handler attribute
type EventAttribute struct {
	Name    string
	Handler interface{}
}

// NewElement creates a new virtual DOM element with mixed arguments
func NewElement(tagType string, args ...interface{}) *Element {
	props := make(map[string]interface{})
	eventHandlers := make(map[string]js.Func)
	children := make([]*Element, 0)

	for _, arg := range args {
		switch v := arg.(type) {
		case Attribute:
			if v.Name != "" { // Skip empty attributes from If() function
				props[v.Name] = v.Value
			}
		case EventAttribute:
			if fn, ok := createEventHandler(v); ok {
				eventHandlers[v.Name] = fn
			}
		case *Element:
			children = append(children, v)
		case string:
			// Text content
			textElement := &Element{
				Type:          "text",
				Props:         map[string]interface{}{"textContent": v},
				Children:      make([]*Element, 0),
				EventHandlers: make(map[string]js.Func),
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

func createEventHandler(event EventAttribute) (js.Func, bool) {
	switch event.Name {
	case "click":
		if handler, ok := event.Handler.(func()); ok {
			return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				handler()
				return nil
			}), true
		}
	case "input":
		if handler, ok := event.Handler.(func(string)); ok {
			return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				handler(args[0].Get("target").Get("value").String())
				return nil
			}), true
		}
	case "change":
		if handler, ok := event.Handler.(func(bool)); ok {
			return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				handler(args[0].Get("target").Get("checked").Bool())
				return nil
			}), true
		}
	case "keydown":
		if handler, ok := event.Handler.(func(string)); ok {
			return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				handler(args[0].Get("key").String())
				return nil
			}), true
		}
	}
	return js.Func{}, false
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
			case "value":
				e.JSElement.Set("value", value)
			case "checked", "autofocus":
				e.JSElement.Set(name, value)
			default:
				e.JSElement.Call("setAttribute", name, fmt.Sprintf("%v", value))
			}
		}

		// Add event listeners
		for event, handler := range e.EventHandlers {
			e.JSElement.Call("addEventListener", event, handler)
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
				case "value":
					e.JSElement.Set("value", newValue)
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

func Placeholder(text string) Attribute {
	return Attribute{Name: "placeholder", Value: text}
}

func Value(value string) Attribute {
	return Attribute{Name: "value", Value: value}
}

func Autofocus(focus bool) Attribute {
	return Attribute{Name: "autofocus", Value: focus}
}

func Type(typeStr string) Attribute {
	return Attribute{Name: "type", Value: typeStr}
}

func Checked(checked bool) Attribute {
	return Attribute{Name: "checked", Value: checked}
}

func On(event string, handler interface{}) EventAttribute {
	return EventAttribute{Name: event, Handler: handler}
}

func OnClick(handler func()) EventAttribute {
	return On("click", handler)
}

func OnInput(handler func(value string)) EventAttribute {
	return On("input", handler)
}

func OnChange(handler func(checked bool)) EventAttribute {
	return On("change", handler)
}

func OnKeyDown(handler func(key string)) EventAttribute {
	return On("keydown", handler)
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

func H3(args ...interface{}) *Element {
	return NewElement("h3", args...)
}

func H4(args ...interface{}) *Element {
	return NewElement("h4", args...)
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

func Checkbox(args ...interface{}) *Element {
	newArgs := append([]interface{}{Type("checkbox")}, args...)
	return NewElement("input", newArgs...)
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

func Label(args ...interface{}) *Element {
	return NewElement("label", args...)
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
