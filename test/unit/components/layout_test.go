//go:build !js || !wasm

package components_test

import (
	"strings"
	"testing"

	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/src/app/components"
)

func TestAppLayoutStructure(t *testing.T) {
	sidebar := dom.Div(dom.Class("sidebar"), dom.Text("Sidebar"))
	content := dom.Div(dom.Class("content"), dom.Text("Content"))

	layout := components.AppLayout(sidebar, content)
	if layout == nil {
		t.Fatal("layout should not be nil")
	}

	html := dom.RenderToHTML(layout)

	// Should contain the app-layout wrapper
	if !strings.Contains(html, "app-layout") {
		t.Error("missing app-layout class")
	}
	// Should contain the sidebar
	if !strings.Contains(html, "sidebar") {
		t.Error("missing sidebar content")
	}
	// Should contain the editor area wrapper
	if !strings.Contains(html, "editor-area") {
		t.Error("missing editor-area class")
	}
}

func TestAppLayoutWithNilSidebar(t *testing.T) {
	content := dom.Div(dom.Text("Content"))
	layout := components.AppLayout(nil, content)
	if layout == nil {
		t.Fatal("layout should handle nil sidebar")
	}
}

func TestAppLayoutWithNilContent(t *testing.T) {
	sidebar := dom.Div(dom.Text("Sidebar"))
	layout := components.AppLayout(sidebar, nil)
	if layout == nil {
		t.Fatal("layout should handle nil content")
	}
}
