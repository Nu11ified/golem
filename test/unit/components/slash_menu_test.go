//go:build !js || !wasm

package components_test

import (
	"strings"
	"testing"

	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/src/app/components"
)

func TestSlashMenuItemCount(t *testing.T) {
	items := components.GetSlashMenuItems()
	if len(items) != 9 {
		t.Errorf("expected 9 menu items, got %d", len(items))
	}
}

func TestSlashMenuItemTypes(t *testing.T) {
	items := components.GetSlashMenuItems()
	expectedTypes := []string{"text", "h1", "h2", "h3", "bullet", "numbered", "toggle", "code", "divider"}
	for i, item := range items {
		if item.BlockType != expectedTypes[i] {
			t.Errorf("item %d: expected type %q, got %q", i, expectedTypes[i], item.BlockType)
		}
	}
}

func TestSlashMenuRendersItems(t *testing.T) {
	menu := components.SlashMenu("")
	if menu == nil {
		t.Fatal("slash menu should not be nil")
	}
	html := dom.RenderToHTML(menu)
	if !strings.Contains(html, "slash-menu") {
		t.Error("missing slash-menu class")
	}
	if !strings.Contains(html, "slash-menu-item") {
		t.Error("missing slash-menu-item class")
	}
}

func TestSlashMenuHasLabels(t *testing.T) {
	menu := components.SlashMenu("")
	html := dom.RenderToHTML(menu)
	if !strings.Contains(html, "Text") {
		t.Error("missing Text label")
	}
	if !strings.Contains(html, "Heading 1") {
		t.Error("missing Heading 1 label")
	}
	if !strings.Contains(html, "Code") {
		t.Error("missing Code label")
	}
	if !strings.Contains(html, "Divider") {
		t.Error("missing Divider label")
	}
}

func TestSlashMenuFilter(t *testing.T) {
	// Filtering with "head" should show heading items
	menu := components.SlashMenu("head")
	html := dom.RenderToHTML(menu)
	if !strings.Contains(html, "Heading") {
		t.Error("filter 'head' should show heading items")
	}
}

func TestSlashMenuFilterNoResults(t *testing.T) {
	menu := components.SlashMenu("zzzzz")
	html := dom.RenderToHTML(menu)
	// Should still render the menu wrapper
	if !strings.Contains(html, "slash-menu") {
		t.Error("even with no results, slash menu wrapper should exist")
	}
}
