//go:build !js || !wasm

package components_test

import (
	"strings"
	"testing"

	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/src/app/components"
	"github.com/Nu11ified/golem/src/app/models"
)

func TestSSRRenderPage(t *testing.T) {
	// Create a workspace with a page
	ws := models.NewWorkspace()
	page := models.NewPage("SSR Test Page", "")
	ws.Pages[page.ID] = page
	ws.RootPageIDs = []string{page.ID}

	// Create blocks for the page
	block := models.NewBlock("text", page.ID)
	block.Content = "Server rendered content"

	es := models.NewEditorState()
	es.ActivePageID = page.ID
	es.Blocks[block.ID] = block
	es.BlockOrder = []string{block.ID}

	// Render the full page
	el := components.RenderFullPage(ws, page, es)
	if el == nil {
		t.Fatal("SSR render should return element")
	}

	html := dom.RenderToHTML(el)

	// Should contain the layout structure
	if !strings.Contains(html, "app-layout") {
		t.Error("missing app-layout class")
	}

	// Should contain the sidebar with page tree
	if !strings.Contains(html, "sidebar") {
		t.Error("missing sidebar")
	}
	if !strings.Contains(html, "SSR Test Page") {
		t.Error("missing page title in sidebar")
	}

	// Should contain the editor
	if !strings.Contains(html, "editor-page") {
		t.Error("missing editor-page")
	}
	if !strings.Contains(html, "Server rendered content") {
		t.Error("missing block content")
	}
}

func TestSSRRenderEmptyPage(t *testing.T) {
	ws := models.NewWorkspace()
	page := models.NewPage("Empty Page", "")
	ws.Pages[page.ID] = page
	ws.RootPageIDs = []string{page.ID}

	es := models.NewEditorState()
	es.ActivePageID = page.ID

	el := components.RenderFullPage(ws, page, es)
	html := dom.RenderToHTML(el)

	// Should still render the layout
	if !strings.Contains(html, "app-layout") {
		t.Error("missing app-layout")
	}
	// Empty page should show placeholder
	if !strings.Contains(html, "block-list-empty") {
		t.Error("empty page should show placeholder")
	}
}

func TestSSRGenerateDocument(t *testing.T) {
	ws := models.NewWorkspace()
	page := models.NewPage("Doc Test", "")
	ws.Pages[page.ID] = page
	ws.RootPageIDs = []string{page.ID}

	es := models.NewEditorState()
	es.ActivePageID = page.ID

	doc := components.GenerateSSRDocument(ws, page, es)

	// Should be a complete HTML document
	if !strings.Contains(doc, "<!DOCTYPE html>") {
		t.Error("missing DOCTYPE")
	}
	if !strings.Contains(doc, "<html") {
		t.Error("missing html tag")
	}
	if !strings.Contains(doc, "Doc Test") {
		t.Error("missing page title")
	}
	// Should include the CSS
	if !strings.Contains(doc, "--bg-sidebar") {
		t.Error("missing CSS variables in document")
	}
	// Should include the initial state script
	if !strings.Contains(doc, "__GOLEM_STATE__") {
		t.Error("missing initial state script tag")
	}
}
