//go:build !js || !wasm

package dom

import (
	"fmt"
	"html"
	"sort"
	"strings"
)

// Stub Element type for non-WASM builds
type Element struct {
	Type             string
	Props            map[string]interface{}
	Children         []*Element
	EventHandlers    map[string]func()
	JSElement        interface{}
	MountCallbacks   []func() func() // returns cleanup function
	UnmountCallbacks []func()
	UpdateCallbacks  []func()
	Cleanups         []func() // stored cleanup functions from mount
	IsMounted        bool
}

// Attribute represents an HTML attribute
type Attribute struct {
	Name  string
	Value interface{}
}

// EventAttribute represents an event handler attribute (server-side stub)
type EventAttribute struct {
	Name    string
	Handler interface{}
}

// LifecycleAttribute represents a lifecycle hook to be applied to an element.
type LifecycleAttribute struct {
	Kind    string      // "mount", "unmount", or "update"
	Handler interface{} // func() func() for mount, func() for unmount/update
}

// selfClosingTags are HTML elements that should not have a closing tag.
var selfClosingTags = map[string]bool{
	"input": true,
	"img":   true,
	"br":    true,
	"hr":    true,
	"meta":  true,
	"link":  true,
}

// booleanAttrs are HTML attributes that should be rendered as bare attributes when true.
var booleanAttrs = map[string]bool{
	"checked":  true,
	"disabled": true,
	"autofocus": true,
	"readonly":  true,
	"required":  true,
	"selected":  true,
	"multiple":  true,
	"hidden":    true,
}

// NewElement creates a new virtual DOM element with mixed arguments
func NewElement(tagType string, args ...interface{}) *Element {
	props := make(map[string]interface{})
	eventHandlers := make(map[string]func())
	children := make([]*Element, 0)
	var mountCallbacks []func() func()
	var unmountCallbacks []func()
	var updateCallbacks []func()

	for _, arg := range args {
		switch v := arg.(type) {
		case Attribute:
			if v.Name != "" { // Skip empty attributes from If() function
				props[v.Name] = v.Value
			}
		case EventAttribute:
			// Store event handlers but they won't be rendered to HTML
			if handler, ok := v.Handler.(func()); ok {
				eventHandlers[v.Name] = handler
			}
			// For non-func() handlers, we just skip them on server side
		case LifecycleAttribute:
			switch v.Kind {
			case "mount":
				if fn, ok := v.Handler.(func() func()); ok {
					mountCallbacks = append(mountCallbacks, fn)
				}
			case "unmount":
				if fn, ok := v.Handler.(func()); ok {
					unmountCallbacks = append(unmountCallbacks, fn)
				}
			case "update":
				if fn, ok := v.Handler.(func()); ok {
					updateCallbacks = append(updateCallbacks, fn)
				}
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
		Type:             tagType,
		Props:            props,
		Children:         children,
		EventHandlers:    eventHandlers,
		MountCallbacks:   mountCallbacks,
		UnmountCallbacks: unmountCallbacks,
		UpdateCallbacks:  updateCallbacks,
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

// RenderToHTML renders an Element tree to an HTML string suitable for
// server-side rendering. Event handlers are omitted. Text content and
// attribute values are HTML-escaped to prevent XSS.
func RenderToHTML(el *Element) string {
	if el == nil {
		return ""
	}
	var buf strings.Builder
	renderElementToHTML(&buf, el)
	return buf.String()
}

// renderElementToHTML recursively renders an element to the builder.
func renderElementToHTML(buf *strings.Builder, el *Element) {
	if el == nil {
		return
	}

	// Handle text nodes
	if el.Type == "text" {
		if tc, ok := el.Props["textContent"]; ok {
			buf.WriteString(html.EscapeString(fmt.Sprintf("%v", tc)))
		}
		return
	}

	// Opening tag
	buf.WriteByte('<')
	buf.WriteString(el.Type)

	// Render attributes from Props in a deterministic order
	renderAttributes(buf, el.Props)

	isSelfClosing := selfClosingTags[el.Type]

	if isSelfClosing {
		buf.WriteString(" />")
		return
	}

	buf.WriteByte('>')

	// Render textContent prop as inner text (not as an attribute)
	if tc, ok := el.Props["textContent"]; ok {
		buf.WriteString(html.EscapeString(fmt.Sprintf("%v", tc)))
	}

	// Render children
	for _, child := range el.Children {
		renderElementToHTML(buf, child)
	}

	// Closing tag
	buf.WriteString("</")
	buf.WriteString(el.Type)
	buf.WriteByte('>')
}

// renderAttributes writes HTML attributes from the props map to the builder.
// It sorts attribute names for deterministic output. It skips textContent
// (rendered as inner text), event handlers, and false boolean attributes.
func renderAttributes(buf *strings.Builder, props map[string]interface{}) {
	if len(props) == 0 {
		return
	}

	// Sort keys for deterministic output
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		value := props[name]

		// Skip textContent — it is rendered as inner text, not an attribute
		if name == "textContent" {
			continue
		}

		// Skip event handler props
		if isEventHandlerProp(name) {
			continue
		}

		// Handle boolean attributes
		if booleanAttrs[name] {
			if bVal, ok := value.(bool); ok {
				if bVal {
					buf.WriteByte(' ')
					buf.WriteString(name)
				}
				// false boolean attributes are omitted entirely
				continue
			}
		}

		// Regular attribute
		buf.WriteByte(' ')
		buf.WriteString(name)
		buf.WriteString(`="`)
		buf.WriteString(html.EscapeString(fmt.Sprintf("%v", value)))
		buf.WriteByte('"')
	}
}

// isEventHandlerProp returns true for props that represent event handlers.
func isEventHandlerProp(name string) bool {
	return strings.HasPrefix(name, "on")
}

// Remove removes the element (stub for non-WASM builds).
func (e *Element) Remove() {
	// Stub implementation: callbacks are stored but not fired in non-WASM builds.
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

func Type(typeStr string) Attribute {
	return Attribute{Name: "type", Value: typeStr}
}

func Checked(checked bool) Attribute {
	return Attribute{Name: "checked", Value: checked}
}

func Autofocus(focus bool) Attribute {
	return Attribute{Name: "autofocus", Value: focus}
}

// OnMount registers a callback that runs after element is added to DOM.
// The callback can return a cleanup function that runs on unmount.
func OnMount(fn func() func()) LifecycleAttribute {
	return LifecycleAttribute{Kind: "mount", Handler: fn}
}

// OnUnmount registers a callback that runs before element is removed from DOM.
func OnUnmount(fn func()) LifecycleAttribute {
	return LifecycleAttribute{Kind: "unmount", Handler: fn}
}

// OnUpdate registers a callback that runs after element re-renders.
func OnUpdate(fn func()) LifecycleAttribute {
	return LifecycleAttribute{Kind: "update", Handler: fn}
}

func Disabled(disabled bool) Attribute {
	return Attribute{Name: "disabled", Value: disabled}
}

// On creates an EventAttribute for a given event name and handler.
func On(event string, handler interface{}) EventAttribute {
	return EventAttribute{Name: event, Handler: handler}
}

// OnClick creates a click event handler (no-op on server, omitted from HTML).
func OnClick(handler func()) EventAttribute {
	return On("click", handler)
}

// OnInput creates an input event handler (no-op on server, omitted from HTML).
func OnInput(handler func(value string)) EventAttribute {
	return On("input", handler)
}

// OnChange creates a change event handler (no-op on server, omitted from HTML).
func OnChange(handler func(checked bool)) EventAttribute {
	return On("change", handler)
}

// OnKeyDown creates a keydown event handler (no-op on server, omitted from HTML).
func OnKeyDown(handler func(key string)) EventAttribute {
	return On("keydown", handler)
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
func H3(args ...interface{}) *Element     { return NewElement("h3", args...) }
func H4(args ...interface{}) *Element     { return NewElement("h4", args...) }
func P(args ...interface{}) *Element      { return NewElement("p", args...) }
func Button(args ...interface{}) *Element { return NewElement("button", args...) }
func Input(args ...interface{}) *Element  { return NewElement("input", args...) }
func Span(args ...interface{}) *Element   { return NewElement("span", args...) }
func A(args ...interface{}) *Element      { return NewElement("a", args...) }
func Img(args ...interface{}) *Element    { return NewElement("img", args...) }
func Ul(args ...interface{}) *Element     { return NewElement("ul", args...) }
func Li(args ...interface{}) *Element     { return NewElement("li", args...) }
func Label(args ...interface{}) *Element  { return NewElement("label", args...) }

// Checkbox creates an input element with type="checkbox".
func Checkbox(args ...interface{}) *Element {
	newArgs := append([]interface{}{Type("checkbox")}, args...)
	return NewElement("input", newArgs...)
}

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

// Style creates an inline style attribute from a map of CSS property→value pairs.
// The resulting attribute renders as style="key: value; key: value;" in HTML.
// An empty map produces a no-op attribute that is omitted from output.
func Style(styles map[string]string) Attribute {
	if len(styles) == 0 {
		return Attribute{Name: "", Value: nil}
	}
	// Sort keys for deterministic output
	keys := make([]string, 0, len(styles))
	for k := range styles {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf strings.Builder
	for _, k := range keys {
		buf.WriteString(k)
		buf.WriteString(": ")
		buf.WriteString(styles[k])
		buf.WriteString("; ")
	}
	result := strings.TrimRight(buf.String(), " ")
	return Attribute{Name: "style", Value: result}
}

// DocumentOptions configures RenderDocument output.
type DocumentOptions struct {
	Title        string
	Lang         string            // default "en"
	Meta         map[string]string // name→content meta tags
	Scripts      []string          // script URLs to include
	Styles       []string          // stylesheet URLs to include
	WasmPath     string            // path to WASM binary
	WasmExecPath string            // path to wasm_exec.js
}

// RenderDocument wraps an element tree in a full HTML document with
// DOCTYPE, html, head, and body tags.
func RenderDocument(body *Element, options DocumentOptions) string {
	lang := options.Lang
	if lang == "" {
		lang = "en"
	}

	var buf strings.Builder
	buf.WriteString("<!DOCTYPE html>\n")
	buf.WriteString(fmt.Sprintf("<html lang=\"%s\">\n", html.EscapeString(lang)))
	buf.WriteString("<head>\n")
	buf.WriteString("<meta charset=\"utf-8\" />\n")

	if options.Title != "" {
		buf.WriteString(fmt.Sprintf("<title>%s</title>\n", html.EscapeString(options.Title)))
	}

	// Meta tags — sorted for deterministic output
	if len(options.Meta) > 0 {
		metaKeys := make([]string, 0, len(options.Meta))
		for k := range options.Meta {
			metaKeys = append(metaKeys, k)
		}
		sort.Strings(metaKeys)
		for _, name := range metaKeys {
			content := options.Meta[name]
			buf.WriteString(fmt.Sprintf("<meta name=\"%s\" content=\"%s\" />\n",
				html.EscapeString(name), html.EscapeString(content)))
		}
	}

	// Stylesheets
	for _, href := range options.Styles {
		buf.WriteString(fmt.Sprintf("<link rel=\"stylesheet\" href=\"%s\" />\n", html.EscapeString(href)))
	}

	// WASM exec script (must come before bootstrap)
	if options.WasmExecPath != "" {
		buf.WriteString(fmt.Sprintf("<script src=\"%s\"></script>\n", html.EscapeString(options.WasmExecPath)))
	}

	// User-specified scripts
	for _, src := range options.Scripts {
		buf.WriteString(fmt.Sprintf("<script src=\"%s\"></script>\n", html.EscapeString(src)))
	}

	buf.WriteString("</head>\n")
	buf.WriteString("<body>\n")

	// Render the body content
	buf.WriteString(RenderToHTML(body))
	buf.WriteByte('\n')

	// WASM bootstrap script (at end of body for DOM readiness)
	if options.WasmPath != "" {
		buf.WriteString("<script>\n")
		buf.WriteString("const go = new Go();\n")
		buf.WriteString(fmt.Sprintf("WebAssembly.instantiateStreaming(fetch(\"%s\"), go.importObject).then((result) => {\n",
			html.EscapeString(options.WasmPath)))
		buf.WriteString("  go.run(result.instance);\n")
		buf.WriteString("});\n")
		buf.WriteString("</script>\n")
	}

	buf.WriteString("</body>\n")
	buf.WriteString("</html>")

	return buf.String()
}

// RenderToHTMLPretty renders an Element tree to an HTML string with
// indentation for human-readable output. The indent parameter specifies
// the string to use for each indentation level (e.g., "  " for two spaces).
func RenderToHTMLPretty(el *Element, indent string) string {
	if el == nil {
		return ""
	}
	var buf strings.Builder
	renderPretty(&buf, el, indent, 0)
	return buf.String()
}

// renderPretty recursively renders an element with indentation.
func renderPretty(buf *strings.Builder, el *Element, indent string, depth int) {
	if el == nil {
		return
	}

	prefix := strings.Repeat(indent, depth)

	// Handle text nodes
	if el.Type == "text" {
		if tc, ok := el.Props["textContent"]; ok {
			buf.WriteString(prefix)
			buf.WriteString(html.EscapeString(fmt.Sprintf("%v", tc)))
		}
		return
	}

	// Check if this element is a simple text-only element (has textContent
	// prop or a single text child, and no other children).
	isSimpleText := false
	simpleText := ""
	if tc, ok := el.Props["textContent"]; ok && len(el.Children) == 0 {
		isSimpleText = true
		simpleText = html.EscapeString(fmt.Sprintf("%v", tc))
	} else if len(el.Children) == 1 && el.Children[0].Type == "text" {
		if tc, ok := el.Children[0].Props["textContent"]; ok {
			_, hasTextProp := el.Props["textContent"]
			if !hasTextProp {
				isSimpleText = true
				simpleText = html.EscapeString(fmt.Sprintf("%v", tc))
			}
		}
	}

	// Opening tag
	buf.WriteString(prefix)
	buf.WriteByte('<')
	buf.WriteString(el.Type)
	renderAttributes(buf, el.Props)

	isSelfClosing := selfClosingTags[el.Type]
	if isSelfClosing {
		buf.WriteString(" />")
		return
	}

	buf.WriteByte('>')

	if isSimpleText {
		// Inline text: <p>Hello</p>
		buf.WriteString(simpleText)
		buf.WriteString("</")
		buf.WriteString(el.Type)
		buf.WriteByte('>')
		return
	}

	// Has complex children
	// Render textContent prop first if present
	if tc, ok := el.Props["textContent"]; ok {
		buf.WriteByte('\n')
		buf.WriteString(strings.Repeat(indent, depth+1))
		buf.WriteString(html.EscapeString(fmt.Sprintf("%v", tc)))
	}

	// Render children
	for _, child := range el.Children {
		buf.WriteByte('\n')
		renderPretty(buf, child, indent, depth+1)
	}

	buf.WriteByte('\n')
	buf.WriteString(prefix)
	buf.WriteString("</")
	buf.WriteString(el.Type)
	buf.WriteByte('>')
}
