//go:build !js || !wasm

package dom_test

import (
	"strings"
	"testing"

	"github.com/Nu11ified/golem/dom"
)

// TestHydrate_StubExists verifies that the Hydrate function is callable
// in non-WASM (stub) builds. In the stub build it is a no-op, but it
// must exist so that application code can reference it without build
// errors regardless of target platform.
func TestHydrate_StubExists(t *testing.T) {
	element := dom.Div(dom.Class("app"), "Hello")

	// Should not panic -- just prints a stub message.
	dom.Hydrate(element, "#app")
}

// TestRenderToHTMLWithIDs_MatchesTreeStructure verifies that IDs are
// assigned in depth-first order matching the tree walk order that
// hydrateNode would use. A tree like:
//
//	div (id=0)
//	  h1 (id=1)
//	  p  (id=2)
//	    span (id=3)
//
// should produce data-golem-id="0" on the div, "1" on h1, "2" on p,
// and "3" on the span.
func TestRenderToHTMLWithIDs_MatchesTreeStructure(t *testing.T) {
	tree := dom.Div(
		dom.H1("Title"),
		dom.P(
			dom.Span("nested"),
		),
	)

	html := dom.RenderToHTMLWithIDs(tree)

	// Verify the IDs appear in the correct depth-first order.
	expectedIDs := []string{
		`data-golem-id="0"`, // div
		`data-golem-id="1"`, // h1
		`data-golem-id="2"`, // p
		`data-golem-id="3"`, // span
	}

	lastIdx := -1
	for _, expected := range expectedIDs {
		idx := strings.Index(html, expected)
		if idx == -1 {
			t.Errorf("expected %q to appear in HTML output, but it was not found.\nHTML: %s", expected, html)
			continue
		}
		if idx <= lastIdx {
			t.Errorf("expected %q (at index %d) to appear after previous ID (at index %d) in depth-first order.\nHTML: %s",
				expected, idx, lastIdx, html)
		}
		lastIdx = idx
	}
}

// TestRenderToHTMLWithIDs_EventHandlersOmitted verifies that event
// handler attributes (like onclick) do NOT appear in the rendered
// HTML. Event handlers are attached during hydration, not serialized
// into the server-rendered HTML.
func TestRenderToHTMLWithIDs_EventHandlersOmitted(t *testing.T) {
	clicked := false
	tree := dom.Div(
		dom.Button(
			dom.OnClick(func() { clicked = true }),
			"Click me",
		),
	)

	html := dom.RenderToHTMLWithIDs(tree)

	// The HTML must not contain any event handler references.
	for _, forbidden := range []string{"onclick", "oninput", "onchange", "onkeydown", "EventHandler", "handler"} {
		if strings.Contains(strings.ToLower(html), strings.ToLower(forbidden)) {
			t.Errorf("HTML should not contain %q, but it does.\nHTML: %s", forbidden, html)
		}
	}

	// Verify the button element itself IS present.
	if !strings.Contains(html, "<button") {
		t.Errorf("expected <button> element in HTML output.\nHTML: %s", html)
	}

	// Verify we haven't accidentally triggered the handler.
	if clicked {
		t.Error("event handler was unexpectedly invoked during rendering")
	}
}

// TestRenderToHTMLWithIDs_ConsistentIDs verifies that the same tree
// produces the same IDs every time (deterministic). This is critical
// for hydration correctness -- the server and client must agree on the
// ID assignment.
func TestRenderToHTMLWithIDs_ConsistentIDs(t *testing.T) {
	buildTree := func() *dom.Element {
		return dom.Div(
			dom.Class("root"),
			dom.H1("Title"),
			dom.Ul(
				dom.Li("Item 1"),
				dom.Li("Item 2"),
				dom.Li("Item 3"),
			),
			dom.P("Footer"),
		)
	}

	first := dom.RenderToHTMLWithIDs(buildTree())
	for i := 0; i < 10; i++ {
		again := dom.RenderToHTMLWithIDs(buildTree())
		if again != first {
			t.Fatalf("RenderToHTMLWithIDs is not deterministic.\nFirst:  %s\nAgain (%d): %s", first, i, again)
		}
	}
}

// TestRenderToHTMLWithIDs_VoidElements verifies that void elements
// like <input> are rendered without a closing tag.
func TestRenderToHTMLWithIDs_VoidElements(t *testing.T) {
	tree := dom.Div(
		dom.Input(dom.Class("field")),
	)

	html := dom.RenderToHTMLWithIDs(tree)

	if strings.Contains(html, "</input>") {
		t.Errorf("void element <input> should not have a closing tag.\nHTML: %s", html)
	}

	if !strings.Contains(html, "<input") {
		t.Errorf("expected <input> element in HTML output.\nHTML: %s", html)
	}
}

// TestRenderToHTMLWithIDs_TextContent verifies that text content is
// rendered as inner text of the element.
func TestRenderToHTMLWithIDs_TextContent(t *testing.T) {
	tree := dom.H1(dom.Text("Hello World"))

	html := dom.RenderToHTMLWithIDs(tree)

	if !strings.Contains(html, "Hello World") {
		t.Errorf("expected 'Hello World' in HTML output.\nHTML: %s", html)
	}

	if !strings.Contains(html, "</h1>") {
		t.Errorf("expected closing </h1> tag.\nHTML: %s", html)
	}
}

// TestRenderToHTMLWithIDs_StringChildren verifies that string children
// (text nodes) are rendered as plain text between element tags, without
// their own data-golem-id.
func TestRenderToHTMLWithIDs_StringChildren(t *testing.T) {
	tree := dom.P("Hello ", dom.Span("world"))

	html := dom.RenderToHTMLWithIDs(tree)

	// The <p> should have data-golem-id="0", text "Hello ", then <span> with id="1".
	if !strings.Contains(html, "Hello ") {
		t.Errorf("expected text 'Hello ' in HTML output.\nHTML: %s", html)
	}

	if !strings.Contains(html, `<span data-golem-id="1"`) {
		t.Errorf("expected span with data-golem-id=\"1\".\nHTML: %s", html)
	}

	// Text nodes should NOT get data-golem-id attributes.
	// Count occurrences of data-golem-id -- should be exactly 2 (p and span).
	count := strings.Count(html, "data-golem-id")
	if count != 2 {
		t.Errorf("expected exactly 2 data-golem-id attributes, got %d.\nHTML: %s", count, html)
	}
}
