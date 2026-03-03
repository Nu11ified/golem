//go:build !js || !wasm

package components_test

import (
	"strings"
	"testing"

	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/src/app/components"
	"github.com/Nu11ified/golem/src/app/models"
)

func TestEditorPageRenders(t *testing.T) {
	ws := models.NewWorkspace()
	page := models.NewPage("Test Page", "")
	ws.Pages[page.ID] = page
	ws.RootPageIDs = []string{page.ID}

	block := models.NewBlock("text", page.ID)
	block.Content = "Hello world"

	es := models.NewEditorState()
	es.ActivePageID = page.ID
	es.Blocks[block.ID] = block
	es.BlockOrder = []string{block.ID}

	el := components.RenderEditorPage(page, es)
	if el == nil {
		t.Fatal("editor page should not be nil")
	}

	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "editor-page") {
		t.Error("missing editor-page class")
	}
	if !strings.Contains(html, "Test Page") {
		t.Error("missing page title")
	}
	if !strings.Contains(html, "Hello world") {
		t.Error("missing block content")
	}
}

func TestBlockListRendersBlocks(t *testing.T) {
	es := models.NewEditorState()
	b1 := models.NewBlock("text", "p1")
	b1.Content = "First block"
	b2 := models.NewBlock("h1", "p1")
	b2.Content = "Heading block"
	es.Blocks[b1.ID] = b1
	es.Blocks[b2.ID] = b2
	es.BlockOrder = []string{b1.ID, b2.ID}

	el := components.BlockList(es)
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "block-list") {
		t.Error("missing block-list class")
	}
	if !strings.Contains(html, "First block") {
		t.Error("missing first block content")
	}
	if !strings.Contains(html, "Heading block") {
		t.Error("missing heading block content")
	}
}

func TestBlockListEmpty(t *testing.T) {
	es := models.NewEditorState()
	el := components.BlockList(es)
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "block-list") {
		t.Error("missing block-list class")
	}
	// Should show a placeholder
	if !strings.Contains(html, "block-list-empty") {
		t.Error("empty block list should have placeholder")
	}
}

func TestEditorPageWithMultipleBlockTypes(t *testing.T) {
	page := models.NewPage("Multi Block Page", "")
	es := models.NewEditorState()
	es.ActivePageID = page.ID

	// Add various block types
	blocks := []*models.Block{
		{ID: "b1", Type: "text", Content: "text content", PageID: page.ID},
		{ID: "b2", Type: "h1", Content: "heading", PageID: page.ID},
		{ID: "b3", Type: "bullet", Content: "bullet item", PageID: page.ID, Children: []string{}, Props: map[string]interface{}{}},
		{ID: "b4", Type: "divider", PageID: page.ID, Children: []string{}, Props: map[string]interface{}{}},
	}
	for _, b := range blocks {
		es.Blocks[b.ID] = b
		es.BlockOrder = append(es.BlockOrder, b.ID)
	}

	el := components.RenderEditorPage(page, es)
	html := dom.RenderToHTML(el)

	if !strings.Contains(html, "text content") {
		t.Error("missing text block")
	}
	if !strings.Contains(html, "heading") {
		t.Error("missing heading block")
	}
	if !strings.Contains(html, "bullet item") {
		t.Error("missing bullet block")
	}
	if !strings.Contains(html, "block-divider") {
		t.Error("missing divider block")
	}
}
