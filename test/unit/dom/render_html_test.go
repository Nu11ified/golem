//go:build !js || !wasm

package dom_test

import (
	"strings"
	"testing"

	"github.com/Nu11ified/golem/dom"
)

func TestRenderToHTML_SimpleDiv(t *testing.T) {
	el := dom.Div()
	html := dom.RenderToHTML(el)
	if html != "<div></div>" {
		t.Errorf("expected <div></div>, got %s", html)
	}
}

func TestRenderToHTML_DivWithClass(t *testing.T) {
	el := dom.Div(dom.Class("container"))
	html := dom.RenderToHTML(el)
	if html != `<div class="container"></div>` {
		t.Errorf("expected <div class=\"container\"></div>, got %s", html)
	}
}

func TestRenderToHTML_DivWithId(t *testing.T) {
	el := dom.Div(dom.Id("app"))
	html := dom.RenderToHTML(el)
	if html != `<div id="app"></div>` {
		t.Errorf("expected <div id=\"app\"></div>, got %s", html)
	}
}

func TestRenderToHTML_NestedElements(t *testing.T) {
	el := dom.Div(
		dom.H1("Hello"),
		dom.P("World"),
	)
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "<h1>Hello</h1>") {
		t.Errorf("expected h1 in output, got %s", html)
	}
	if !strings.Contains(html, "<p>World</p>") {
		t.Errorf("expected p in output, got %s", html)
	}
}

func TestRenderToHTML_TextContent(t *testing.T) {
	el := dom.Span(dom.Text("hello"))
	html := dom.RenderToHTML(el)
	if html != "<span>hello</span>" {
		t.Errorf("expected <span>hello</span>, got %s", html)
	}
}

func TestRenderToHTML_StringArgTextContent(t *testing.T) {
	el := dom.P("some text")
	html := dom.RenderToHTML(el)
	if html != "<p>some text</p>" {
		t.Errorf("expected <p>some text</p>, got %s", html)
	}
}

func TestRenderToHTML_InputSelfClosing(t *testing.T) {
	el := dom.Input(dom.Type("text"), dom.Placeholder("Enter name"))
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "type=\"text\"") {
		t.Errorf("expected type attr, got %s", html)
	}
	if !strings.Contains(html, "placeholder=\"Enter name\"") {
		t.Errorf("expected placeholder attr, got %s", html)
	}
	if !strings.HasSuffix(html, "/>") {
		t.Errorf("expected self-closing tag for input, got %s", html)
	}
	if strings.Contains(html, "</input>") {
		t.Errorf("input should not have closing tag, got %s", html)
	}
}

func TestRenderToHTML_ImgSelfClosing(t *testing.T) {
	el := dom.Img(dom.Attribute{Name: "src", Value: "pic.png"}, dom.Attribute{Name: "alt", Value: "A picture"})
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "src=\"pic.png\"") {
		t.Errorf("expected src attr, got %s", html)
	}
	if !strings.Contains(html, "alt=\"A picture\"") {
		t.Errorf("expected alt attr, got %s", html)
	}
	if !strings.HasSuffix(html, "/>") {
		t.Errorf("expected self-closing tag for img, got %s", html)
	}
}

func TestRenderToHTML_BrSelfClosing(t *testing.T) {
	el := dom.NewElement("br")
	html := dom.RenderToHTML(el)
	if html != "<br />" {
		t.Errorf("expected <br />, got %s", html)
	}
}

func TestRenderToHTML_HrSelfClosing(t *testing.T) {
	el := dom.NewElement("hr")
	html := dom.RenderToHTML(el)
	if html != "<hr />" {
		t.Errorf("expected <hr />, got %s", html)
	}
}

func TestRenderToHTML_BooleanAttributes(t *testing.T) {
	el := dom.Input(dom.Checked(true), dom.Disabled(true))
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "checked") {
		t.Errorf("expected checked attr, got %s", html)
	}
	if !strings.Contains(html, "disabled") {
		t.Errorf("expected disabled attr, got %s", html)
	}
	// Boolean true should be bare attribute, not checked="true"
	if strings.Contains(html, `checked="true"`) {
		t.Errorf("boolean attr should be bare, not checked=\"true\", got %s", html)
	}
	if strings.Contains(html, `disabled="true"`) {
		t.Errorf("boolean attr should be bare, not disabled=\"true\", got %s", html)
	}
}

func TestRenderToHTML_BooleanAttributeFalse(t *testing.T) {
	el := dom.Input(dom.Checked(false), dom.Disabled(false))
	html := dom.RenderToHTML(el)
	if strings.Contains(html, "checked") {
		t.Errorf("false boolean attr should be omitted, got %s", html)
	}
	if strings.Contains(html, "disabled") {
		t.Errorf("false boolean attr should be omitted, got %s", html)
	}
}

func TestRenderToHTML_AutofocusAttribute(t *testing.T) {
	el := dom.Input(dom.Autofocus(true))
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "autofocus") {
		t.Errorf("expected autofocus attr, got %s", html)
	}
	if strings.Contains(html, `autofocus="true"`) {
		t.Errorf("boolean attr should be bare, not autofocus=\"true\", got %s", html)
	}
}

func TestRenderToHTML_ComplexTree(t *testing.T) {
	el := dom.Div(dom.Class("app"),
		dom.Div(dom.Class("sidebar"),
			dom.Ul(
				dom.Li("Item 1"),
				dom.Li("Item 2"),
			),
		),
		dom.Div(dom.Class("content"),
			dom.H1("Title"),
			dom.P("Paragraph text"),
		),
	)
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, `class="app"`) {
		t.Errorf("expected app class, got %s", html)
	}
	if !strings.Contains(html, `class="sidebar"`) {
		t.Errorf("expected sidebar class, got %s", html)
	}
	if !strings.Contains(html, "<li>Item 1</li>") {
		t.Errorf("expected list item, got %s", html)
	}
	if !strings.Contains(html, "<li>Item 2</li>") {
		t.Errorf("expected second list item, got %s", html)
	}
	if !strings.Contains(html, `class="content"`) {
		t.Errorf("expected content class, got %s", html)
	}
	if !strings.Contains(html, "<h1>Title</h1>") {
		t.Errorf("expected h1 title, got %s", html)
	}
	if !strings.Contains(html, "<p>Paragraph text</p>") {
		t.Errorf("expected paragraph, got %s", html)
	}
}

func TestRenderToHTML_EventHandlersOmitted(t *testing.T) {
	el := dom.Button(dom.Text("Click"), dom.OnClick(func() {}))
	html := dom.RenderToHTML(el)
	if strings.Contains(html, "onclick") {
		t.Errorf("event handlers should not appear in HTML, got %s", html)
	}
	if strings.Contains(html, "click") && !strings.Contains(html, "Click") {
		t.Errorf("event handler name should not appear in HTML, got %s", html)
	}
	if !strings.Contains(html, "Click") {
		t.Errorf("expected button text, got %s", html)
	}
}

func TestRenderToHTML_OnInputOmitted(t *testing.T) {
	el := dom.Input(dom.OnInput(func(v string) {}))
	html := dom.RenderToHTML(el)
	if strings.Contains(html, "oninput") || strings.Contains(html, "input") && !strings.HasPrefix(html, "<input") {
		t.Errorf("event handlers should not appear in HTML, got %s", html)
	}
}

func TestRenderToHTML_OnChangeOmitted(t *testing.T) {
	el := dom.Input(dom.OnChange(func(v bool) {}))
	html := dom.RenderToHTML(el)
	if strings.Contains(html, "onchange") {
		t.Errorf("event handlers should not appear in HTML, got %s", html)
	}
}

func TestRenderToHTML_OnKeyDownOmitted(t *testing.T) {
	el := dom.Input(dom.OnKeyDown(func(k string) {}))
	html := dom.RenderToHTML(el)
	if strings.Contains(html, "onkeydown") {
		t.Errorf("event handlers should not appear in HTML, got %s", html)
	}
}

func TestRenderToHTML_HTMLEscapingText(t *testing.T) {
	el := dom.P("<script>alert('xss')</script>")
	html := dom.RenderToHTML(el)
	if strings.Contains(html, "<script>") {
		t.Errorf("text content should be HTML-escaped, got %s", html)
	}
	if !strings.Contains(html, "&lt;script&gt;") {
		t.Errorf("expected escaped script tag, got %s", html)
	}
}

func TestRenderToHTML_HTMLEscapingTextAttr(t *testing.T) {
	el := dom.Span(dom.Text("<b>bold</b>"))
	html := dom.RenderToHTML(el)
	if strings.Contains(html, "<b>bold</b>") {
		t.Errorf("text from Text() should be HTML-escaped, got %s", html)
	}
	if !strings.Contains(html, "&lt;b&gt;bold&lt;/b&gt;") {
		t.Errorf("expected escaped bold tag, got %s", html)
	}
}

func TestRenderToHTML_HTMLEscapingAttributeValue(t *testing.T) {
	el := dom.Div(dom.Class(`" onmouseover="alert(1)`))
	html := dom.RenderToHTML(el)
	// The quotes inside the attribute value must be escaped so the browser
	// treats the entire string as a single attribute value, not as an
	// attribute boundary that would allow attribute injection.
	if strings.Contains(html, `class="" onmouseover=`) {
		t.Errorf("attribute value should be escaped to prevent XSS, got %s", html)
	}
	// The escaped form should use &#34; for the embedded quotes
	if !strings.Contains(html, `&#34;`) {
		t.Errorf("expected HTML-escaped quotes in attribute value, got %s", html)
	}
}

func TestRenderToHTML_DataGolemId(t *testing.T) {
	el := dom.Div(dom.Attribute{Name: "data-golem-id", Value: "c1"})
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, `data-golem-id="c1"`) {
		t.Errorf("expected data-golem-id attribute, got %s", html)
	}
}

func TestRenderToHTML_Checkbox(t *testing.T) {
	el := dom.Checkbox(dom.Checked(true))
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, `type="checkbox"`) {
		t.Errorf("expected type=checkbox, got %s", html)
	}
	if !strings.Contains(html, "checked") {
		t.Errorf("expected checked attr, got %s", html)
	}
}

func TestRenderToHTML_Label(t *testing.T) {
	el := dom.Label("Username")
	html := dom.RenderToHTML(el)
	if html != "<label>Username</label>" {
		t.Errorf("expected <label>Username</label>, got %s", html)
	}
}

func TestRenderToHTML_H3(t *testing.T) {
	el := dom.H3("Subtitle")
	html := dom.RenderToHTML(el)
	if html != "<h3>Subtitle</h3>" {
		t.Errorf("expected <h3>Subtitle</h3>, got %s", html)
	}
}

func TestRenderToHTML_H4(t *testing.T) {
	el := dom.H4("Section")
	html := dom.RenderToHTML(el)
	if html != "<h4>Section</h4>" {
		t.Errorf("expected <h4>Section</h4>, got %s", html)
	}
}

func TestRenderToHTML_ValueAttribute(t *testing.T) {
	el := dom.Input(dom.Type("text"), dom.Value("hello"))
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, `value="hello"`) {
		t.Errorf("expected value attr, got %s", html)
	}
}

func TestRenderToHTML_PlaceholderAttribute(t *testing.T) {
	el := dom.Input(dom.Placeholder("Enter..."))
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, `placeholder="Enter..."`) {
		t.Errorf("expected placeholder attr, got %s", html)
	}
}

func TestRenderToHTML_MixedChildrenAndTextContent(t *testing.T) {
	// When textContent prop is set AND there are children,
	// textContent should render as text before children
	el := dom.Div(dom.Text("prefix"), dom.Span("inner"))
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "prefix") {
		t.Errorf("expected textContent in output, got %s", html)
	}
	if !strings.Contains(html, "<span>inner</span>") {
		t.Errorf("expected child span in output, got %s", html)
	}
}

func TestRenderToHTML_NilElement(t *testing.T) {
	html := dom.RenderToHTML(nil)
	if html != "" {
		t.Errorf("expected empty string for nil element, got %s", html)
	}
}

func TestRenderToHTML_AnchorWithHref(t *testing.T) {
	el := dom.A(dom.Attribute{Name: "href", Value: "/about"}, "About Us")
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, `href="/about"`) {
		t.Errorf("expected href attribute, got %s", html)
	}
	if !strings.Contains(html, "About Us") {
		t.Errorf("expected link text, got %s", html)
	}
}

func TestRenderToHTML_MetaSelfClosing(t *testing.T) {
	el := dom.NewElement("meta", dom.Attribute{Name: "charset", Value: "utf-8"})
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, `charset="utf-8"`) {
		t.Errorf("expected charset attr, got %s", html)
	}
	if !strings.HasSuffix(html, "/>") {
		t.Errorf("expected self-closing tag for meta, got %s", html)
	}
}

func TestRenderToHTML_LinkSelfClosing(t *testing.T) {
	el := dom.NewElement("link", dom.Attribute{Name: "rel", Value: "stylesheet"}, dom.Attribute{Name: "href", Value: "style.css"})
	html := dom.RenderToHTML(el)
	if !strings.HasSuffix(html, "/>") {
		t.Errorf("expected self-closing tag for link, got %s", html)
	}
}

func TestRenderToHTML_DeeplyNested(t *testing.T) {
	el := dom.Div(
		dom.Div(
			dom.Div(
				dom.Div(
					dom.P("deep"),
				),
			),
		),
	)
	html := dom.RenderToHTML(el)
	expected := "<div><div><div><div><p>deep</p></div></div></div></div>"
	if html != expected {
		t.Errorf("expected %s, got %s", expected, html)
	}
}

// =====================================================================
// Task 10: Complete Server-Side Element Rendering — New tests
// =====================================================================

func TestRenderToHTML_StyleAttribute(t *testing.T) {
	el := dom.Div(dom.Style(map[string]string{"color": "red", "font-size": "16px"}))
	html := dom.RenderToHTML(el)
	// Style should be rendered as a style attribute. Since map iteration order
	// is non-deterministic, check for both orderings.
	if !strings.Contains(html, `style="`) {
		t.Fatalf("expected style attribute, got %s", html)
	}
	if !strings.Contains(html, "color: red;") {
		t.Errorf("expected color: red; in style, got %s", html)
	}
	if !strings.Contains(html, "font-size: 16px;") {
		t.Errorf("expected font-size: 16px; in style, got %s", html)
	}
	if !strings.HasPrefix(html, "<div") || !strings.HasSuffix(html, "</div>") {
		t.Errorf("expected div wrapper, got %s", html)
	}
}

func TestRenderToHTML_StyleEscaping(t *testing.T) {
	// Style values containing HTML-special characters should be escaped
	el := dom.Div(dom.Style(map[string]string{
		"content": `"<script>alert('xss')</script>"`,
	}))
	html := dom.RenderToHTML(el)
	if strings.Contains(html, "<script>") {
		t.Errorf("style values should be HTML-escaped, got %s", html)
	}
	if !strings.Contains(html, "&lt;script&gt;") {
		t.Errorf("expected escaped script tag in style value, got %s", html)
	}
}

func TestRenderToHTML_DataGolemIdIncrementing(t *testing.T) {
	el := dom.Div(
		dom.Span("one"),
		dom.P("two"),
		dom.Div(
			dom.Span("three"),
		),
	)
	html := dom.RenderToHTMLWithIDs(el)

	// The outer div should get data-golem-id="0"
	if !strings.Contains(html, `data-golem-id="0"`) {
		t.Errorf("expected data-golem-id=0 on outer div, got %s", html)
	}
	// Span "one" should get data-golem-id="1"
	if !strings.Contains(html, `data-golem-id="1"`) {
		t.Errorf("expected data-golem-id=1 on first child, got %s", html)
	}
	// P "two" should get data-golem-id="2"
	if !strings.Contains(html, `data-golem-id="2"`) {
		t.Errorf("expected data-golem-id=2 on second child, got %s", html)
	}
	// Inner div should get data-golem-id="3"
	if !strings.Contains(html, `data-golem-id="3"`) {
		t.Errorf("expected data-golem-id=3 on inner div, got %s", html)
	}
	// Inner span should get data-golem-id="4"
	if !strings.Contains(html, `data-golem-id="4"`) {
		t.Errorf("expected data-golem-id=4 on inner span, got %s", html)
	}
}

func TestRenderDocument_BasicHTML(t *testing.T) {
	body := dom.Div(dom.Id("app"), dom.P("Hello, World!"))
	html := dom.RenderDocument(body, dom.DocumentOptions{
		Title: "My App",
	})

	if !strings.HasPrefix(html, "<!DOCTYPE html>") {
		t.Errorf("expected DOCTYPE at start, got %s", html[:50])
	}
	if !strings.Contains(html, "<html lang=\"en\">") {
		t.Errorf("expected html tag with lang=en, got %s", html)
	}
	if !strings.Contains(html, "<head>") {
		t.Errorf("expected head tag, got %s", html)
	}
	if !strings.Contains(html, "<title>My App</title>") {
		t.Errorf("expected title tag, got %s", html)
	}
	if !strings.Contains(html, "<body>") {
		t.Errorf("expected body tag, got %s", html)
	}
	if !strings.Contains(html, "</html>") {
		t.Errorf("expected closing html tag, got %s", html)
	}
	if !strings.Contains(html, `id="app"`) {
		t.Errorf("expected app div in body, got %s", html)
	}
	if !strings.Contains(html, "Hello, World!") {
		t.Errorf("expected body content, got %s", html)
	}
}

func TestRenderDocument_CustomLang(t *testing.T) {
	body := dom.Div()
	html := dom.RenderDocument(body, dom.DocumentOptions{
		Title: "Test",
		Lang:  "fr",
	})
	if !strings.Contains(html, `<html lang="fr">`) {
		t.Errorf("expected lang=fr, got %s", html)
	}
}

func TestRenderDocument_WithMeta(t *testing.T) {
	body := dom.Div()
	html := dom.RenderDocument(body, dom.DocumentOptions{
		Title: "Test",
		Meta: map[string]string{
			"description": "A test page",
			"viewport":    "width=device-width, initial-scale=1",
		},
	})
	if !strings.Contains(html, `<meta name="description" content="A test page"`) {
		t.Errorf("expected description meta tag, got %s", html)
	}
	if !strings.Contains(html, `<meta name="viewport" content="width=device-width, initial-scale=1"`) {
		t.Errorf("expected viewport meta tag, got %s", html)
	}
}

func TestRenderDocument_WithScripts(t *testing.T) {
	body := dom.Div()
	html := dom.RenderDocument(body, dom.DocumentOptions{
		Title:   "Test",
		Scripts: []string{"app.js", "vendor.js"},
	})
	if !strings.Contains(html, `<script src="app.js"></script>`) {
		t.Errorf("expected app.js script tag, got %s", html)
	}
	if !strings.Contains(html, `<script src="vendor.js"></script>`) {
		t.Errorf("expected vendor.js script tag, got %s", html)
	}
}

func TestRenderDocument_WithStyles(t *testing.T) {
	body := dom.Div()
	html := dom.RenderDocument(body, dom.DocumentOptions{
		Title:  "Test",
		Styles: []string{"style.css", "theme.css"},
	})
	if !strings.Contains(html, `<link rel="stylesheet" href="style.css"`) {
		t.Errorf("expected style.css link tag, got %s", html)
	}
	if !strings.Contains(html, `<link rel="stylesheet" href="theme.css"`) {
		t.Errorf("expected theme.css link tag, got %s", html)
	}
}

func TestRenderDocument_WithWASM(t *testing.T) {
	body := dom.Div(dom.Id("app"))
	html := dom.RenderDocument(body, dom.DocumentOptions{
		Title:        "WASM App",
		WasmPath:     "app.wasm",
		WasmExecPath: "wasm_exec.js",
	})

	// Should include wasm_exec.js
	if !strings.Contains(html, `<script src="wasm_exec.js"></script>`) {
		t.Errorf("expected wasm_exec.js script tag, got %s", html)
	}
	// Should include WASM bootstrap code
	if !strings.Contains(html, "WebAssembly.instantiateStreaming") {
		t.Errorf("expected WebAssembly.instantiateStreaming in output, got %s", html)
	}
	if !strings.Contains(html, "app.wasm") {
		t.Errorf("expected app.wasm path in WASM bootstrap, got %s", html)
	}
	if !strings.Contains(html, "new Go()") {
		t.Errorf("expected Go runtime instantiation in WASM bootstrap, got %s", html)
	}
}

func TestRenderToHTMLPretty_Indentation(t *testing.T) {
	el := dom.Div(dom.Class("outer"),
		dom.P("Hello"),
		dom.Div(dom.Class("inner"),
			dom.Span("World"),
		),
	)
	html := dom.RenderToHTMLPretty(el, "  ")

	lines := strings.Split(html, "\n")

	// First line should be the opening div (no indent at top level)
	if !strings.HasPrefix(lines[0], "<div") {
		t.Errorf("expected first line to start with <div, got %q", lines[0])
	}

	// Check that child elements are indented
	foundIndentedP := false
	foundDoubleIndentedSpan := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(line, "  ") && strings.HasPrefix(trimmed, "<p>") {
			foundIndentedP = true
		}
		if strings.HasPrefix(line, "    ") && strings.HasPrefix(trimmed, "<span>") {
			foundDoubleIndentedSpan = true
		}
	}

	if !foundIndentedP {
		t.Errorf("expected indented <p> element, got:\n%s", html)
	}
	if !foundDoubleIndentedSpan {
		t.Errorf("expected double-indented <span> element, got:\n%s", html)
	}
}

func TestRenderToHTMLPretty_SelfClosing(t *testing.T) {
	el := dom.Div(
		dom.Input(dom.Type("text")),
	)
	html := dom.RenderToHTMLPretty(el, "  ")

	if !strings.Contains(html, "  <input") {
		t.Errorf("expected indented input, got:\n%s", html)
	}
	if !strings.Contains(html, "/>") {
		t.Errorf("expected self-closing input tag, got:\n%s", html)
	}
}

func TestRenderToHTMLPretty_TextNoExtraLine(t *testing.T) {
	// For elements with only text content and no child elements,
	// the text should be inline: <p>Hello</p>
	el := dom.P("Hello")
	html := dom.RenderToHTMLPretty(el, "  ")
	if html != "<p>Hello</p>" {
		t.Errorf("expected inline text element, got %q", html)
	}
}

func TestStyle_EmptyMap(t *testing.T) {
	el := dom.Div(dom.Style(map[string]string{}))
	html := dom.RenderToHTML(el)
	// Empty style map should not produce a style attribute
	if strings.Contains(html, "style") {
		t.Errorf("empty style map should not produce style attribute, got %s", html)
	}
}

func TestRenderToHTMLWithIDs_SelfClosing(t *testing.T) {
	el := dom.Div(
		dom.Input(dom.Type("text")),
	)
	html := dom.RenderToHTMLWithIDs(el)
	if !strings.Contains(html, `data-golem-id="0"`) {
		t.Errorf("expected data-golem-id=0 on div, got %s", html)
	}
	if !strings.Contains(html, `data-golem-id="1"`) {
		t.Errorf("expected data-golem-id=1 on input, got %s", html)
	}
}
