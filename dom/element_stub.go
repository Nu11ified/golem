//go:build !js || !wasm

package dom

import (
	"fmt"
	"sort"
	"strings"
)

// Stub Element type for non-WASM builds
type Element struct {
	Type          string
	Props         map[string]interface{}
	Children      []*Element
	EventHandlers map[string]func()
	JSElement     interface{}
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

// Render returns a placeholder for non-WASM builds
func (e *Element) Render() interface{} {
	return fmt.Sprintf("<%s>", e.Type)
}

// Update updates the element with new props
func (e *Element) Update(newProps map[string]interface{}) {
	// Stub implementation for non-WASM builds
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
func Div(args ...interface{}) *Element    { return NewElement("div", args...) }
func H1(args ...interface{}) *Element     { return NewElement("h1", args...) }
func H2(args ...interface{}) *Element     { return NewElement("h2", args...) }
func P(args ...interface{}) *Element      { return NewElement("p", args...) }
func Button(args ...interface{}) *Element { return NewElement("button", args...) }
func Input(args ...interface{}) *Element  { return NewElement("input", args...) }
func Span(args ...interface{}) *Element   { return NewElement("span", args...) }
func A(args ...interface{}) *Element      { return NewElement("a", args...) }
func Img(args ...interface{}) *Element    { return NewElement("img", args...) }
func Ul(args ...interface{}) *Element     { return NewElement("ul", args...) }
func Li(args ...interface{}) *Element     { return NewElement("li", args...) }

// Render renders an element tree to a target selector (stub)
func Render(element *Element, selector string) {
	fmt.Printf("Rendering %s to %s (stub)\n", element.Type, selector)
}

// Hydrate attaches event handlers from the virtual element tree to
// existing server-rendered DOM nodes, making the page interactive
// without re-rendering. (stub for non-WASM builds)
func Hydrate(element *Element, selector string) {
	fmt.Printf("Hydrate: %s to %s (stub)\n", element.Type, selector)
}

// RenderToHTMLWithIDs renders an element tree to an HTML string with
// data-golem-id attributes for hydration matching. IDs are assigned in
// depth-first order so that the hydration walk can match virtual nodes
// to real DOM nodes deterministically.
func RenderToHTMLWithIDs(element *Element) string {
	var counter int
	return renderNodeWithIDs(element, &counter)
}

// voidElements is the set of HTML elements that must not have a closing tag.
var voidElements = map[string]bool{
	"area": true, "base": true, "br": true, "col": true,
	"embed": true, "hr": true, "img": true, "input": true,
	"link": true, "meta": true, "source": true, "track": true,
	"wbr": true,
}

func renderNodeWithIDs(element *Element, counter *int) string {
	if element == nil {
		return ""
	}

	// Text nodes render as plain text (no wrapper element, no ID).
	if element.Type == "text" {
		if tc, ok := element.Props["textContent"]; ok {
			return fmt.Sprintf("%v", tc)
		}
		return ""
	}

	id := *counter
	*counter++

	var sb strings.Builder
	sb.WriteString("<")
	sb.WriteString(element.Type)

	// data-golem-id attribute first for consistency
	sb.WriteString(fmt.Sprintf(` data-golem-id="%d"`, id))

	// Render props as attributes (sorted for determinism).
	// Event handlers are intentionally omitted -- they are attached
	// during hydration, not serialized into the HTML.
	keys := make([]string, 0, len(element.Props))
	for k := range element.Props {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		value := element.Props[name]
		switch name {
		case "textContent":
			// textContent is rendered as inner text, not an attribute
			continue
		case "checked":
			if b, ok := value.(bool); ok && b {
				sb.WriteString(" checked")
			}
		case "disabled":
			if b, ok := value.(bool); ok && b {
				sb.WriteString(" disabled")
			}
		case "autofocus":
			if b, ok := value.(bool); ok && b {
				sb.WriteString(" autofocus")
			}
		default:
			sb.WriteString(fmt.Sprintf(` %s="%v"`, name, value))
		}
	}

	sb.WriteString(">")

	// Void elements have no closing tag and no children.
	if voidElements[element.Type] {
		return sb.String()
	}

	// If there is a textContent prop and no children, render it as inner text.
	if tc, ok := element.Props["textContent"]; ok && len(element.Children) == 0 {
		sb.WriteString(fmt.Sprintf("%v", tc))
	}

	// Render children recursively.
	for _, child := range element.Children {
		sb.WriteString(renderNodeWithIDs(child, counter))
	}

	sb.WriteString(fmt.Sprintf("</%s>", element.Type))
	return sb.String()
}

// Alert shows a browser alert (stub)
func Alert(message string) {
	fmt.Printf("Alert: %s (stub)\n", message)
}
