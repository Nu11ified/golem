//go:build !js || !wasm

package dom

import "fmt"

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

// Alert shows a browser alert (stub)
func Alert(message string) {
	fmt.Printf("Alert: %s (stub)\n", message)
}

// voidElements are HTML elements that cannot have children and must not have a closing tag
var voidElements = map[string]bool{
	"area": true, "base": true, "br": true, "col": true,
	"embed": true, "hr": true, "img": true, "input": true,
	"link": true, "meta": true, "param": true, "source": true,
	"track": true, "wbr": true,
}

// RenderToHTML converts an Element tree to an HTML string.
func RenderToHTML(e *Element) string {
	if e == nil {
		return ""
	}
	return renderElementToHTML(e, false, nil)
}

// RenderToHTMLWithIDs converts an Element tree to an HTML string with
// unique data-golem-id attributes for hydration.
func RenderToHTMLWithIDs(e *Element) string {
	if e == nil {
		return ""
	}
	counter := 0
	return renderElementToHTML(e, true, &counter)
}

// renderElementToHTML is the recursive implementation for HTML rendering.
func renderElementToHTML(e *Element, withIDs bool, counter *int) string {
	if e == nil {
		return ""
	}

	// Handle text nodes
	if e.Type == "text" {
		if tc, ok := e.Props["textContent"]; ok {
			return htmlEscape(fmt.Sprintf("%v", tc))
		}
		return ""
	}

	// Build opening tag
	html := "<" + e.Type

	// Add hydration ID if requested
	if withIDs && counter != nil {
		html += fmt.Sprintf(` data-golem-id="%d"`, *counter)
		*counter++
	}

	// Add attributes from props (skip event-like and internal props)
	// Sort keys for deterministic output
	for _, key := range sortedKeys(e.Props) {
		val := e.Props[key]
		switch key {
		case "textContent", "onclick":
			// Skip these - textContent is handled as children, onclick is an event
			continue
		case "class":
			html += fmt.Sprintf(` class="%s"`, htmlAttrEscape(fmt.Sprintf("%v", val)))
		case "id":
			html += fmt.Sprintf(` id="%s"`, htmlAttrEscape(fmt.Sprintf("%v", val)))
		case "checked", "disabled", "autofocus":
			if b, ok := val.(bool); ok && b {
				html += " " + key
			}
		default:
			html += fmt.Sprintf(` %s="%s"`, key, htmlAttrEscape(fmt.Sprintf("%v", val)))
		}
	}

	// Void elements
	if voidElements[e.Type] {
		html += ">"
		return html
	}

	html += ">"

	// If there is textContent and no children, render as text
	if tc, ok := e.Props["textContent"]; ok && len(e.Children) == 0 {
		html += htmlEscape(fmt.Sprintf("%v", tc))
	}

	// Render children
	for _, child := range e.Children {
		html += renderElementToHTML(child, withIDs, counter)
	}

	html += "</" + e.Type + ">"
	return html
}

// sortedKeys returns map keys sorted alphabetically for deterministic output.
func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// Simple insertion sort for small maps
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}

// htmlEscape escapes special HTML characters in text content.
func htmlEscape(s string) string {
	result := ""
	for _, c := range s {
		switch c {
		case '&':
			result += "&amp;"
		case '<':
			result += "&lt;"
		case '>':
			result += "&gt;"
		default:
			result += string(c)
		}
	}
	return result
}

// htmlAttrEscape escapes special characters in HTML attribute values.
func htmlAttrEscape(s string) string {
	result := ""
	for _, c := range s {
		switch c {
		case '&':
			result += "&amp;"
		case '<':
			result += "&lt;"
		case '>':
			result += "&gt;"
		case '"':
			result += "&quot;"
		default:
			result += string(c)
		}
	}
	return result
}

// DocumentOptions configures the HTML document wrapper.
type DocumentOptions struct {
	Title     string
	Lang      string
	Meta      []MetaTag
	Styles    []string
	BodyAttrs string
	HeadExtra string
}

// MetaTag represents an HTML meta tag.
type MetaTag struct {
	Name    string
	Content string
	// For http-equiv, charset, etc.
	Attrs map[string]string
}

// RenderDocument wraps HTML content in a full HTML document.
func RenderDocument(bodyHTML string, opts DocumentOptions) string {
	lang := opts.Lang
	if lang == "" {
		lang = "en"
	}

	doc := "<!DOCTYPE html>\n"
	doc += fmt.Sprintf("<html lang=\"%s\">\n", htmlAttrEscape(lang))
	doc += "<head>\n"
	doc += "<meta charset=\"UTF-8\">\n"
	doc += "<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n"

	if opts.Title != "" {
		doc += fmt.Sprintf("<title>%s</title>\n", htmlEscape(opts.Title))
	}

	for _, meta := range opts.Meta {
		if meta.Attrs != nil {
			doc += "<meta"
			for k, v := range meta.Attrs {
				doc += fmt.Sprintf(` %s="%s"`, k, htmlAttrEscape(v))
			}
			doc += ">\n"
		} else if meta.Name != "" {
			doc += fmt.Sprintf("<meta name=\"%s\" content=\"%s\">\n",
				htmlAttrEscape(meta.Name), htmlAttrEscape(meta.Content))
		}
	}

	for _, style := range opts.Styles {
		doc += fmt.Sprintf("<link rel=\"stylesheet\" href=\"%s\">\n", htmlAttrEscape(style))
	}

	if opts.HeadExtra != "" {
		doc += opts.HeadExtra + "\n"
	}

	doc += "</head>\n"

	if opts.BodyAttrs != "" {
		doc += fmt.Sprintf("<body %s>\n", opts.BodyAttrs)
	} else {
		doc += "<body>\n"
	}

	doc += bodyHTML + "\n"

	doc += "</body>\n"
	doc += "</html>"

	return doc
}
