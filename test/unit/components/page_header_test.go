//go:build !js || !wasm

package components_test

import (
	"strings"
	"testing"

	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/src/app/components"
	"github.com/Nu11ified/golem/src/app/models"
)

func TestPageHeaderRendersTitle(t *testing.T) {
	page := models.NewPage("My Test Page", "")
	header := components.PageHeader(page)
	if header == nil {
		t.Fatal("page header should not be nil")
	}

	html := dom.RenderToHTML(header)
	if !strings.Contains(html, "My Test Page") {
		t.Error("missing page title in header")
	}
}

func TestPageHeaderRendersIcon(t *testing.T) {
	page := models.NewPage("Test", "")
	page.Icon = "🚀"
	header := components.PageHeader(page)
	html := dom.RenderToHTML(header)
	if !strings.Contains(html, "🚀") {
		t.Error("missing icon in header")
	}
}

func TestPageHeaderHasCorrectClasses(t *testing.T) {
	page := models.NewPage("Test", "")
	header := components.PageHeader(page)
	html := dom.RenderToHTML(header)
	if !strings.Contains(html, "page-header") {
		t.Error("missing page-header class")
	}
	if !strings.Contains(html, "page-header-icon") {
		t.Error("missing page-header-icon class")
	}
	if !strings.Contains(html, "page-header-title") {
		t.Error("missing page-header-title class")
	}
}

func TestPageHeaderUntitled(t *testing.T) {
	page := models.NewPage("", "")
	page.Title = ""
	header := components.PageHeader(page)
	html := dom.RenderToHTML(header)
	if !strings.Contains(html, "Untitled") {
		t.Error("empty title should show 'Untitled' placeholder")
	}
}

func TestEmojiPickerItems(t *testing.T) {
	picker := components.EmojiPicker()
	if picker == nil {
		t.Fatal("emoji picker should not be nil")
	}
	html := dom.RenderToHTML(picker)
	if !strings.Contains(html, "emoji-picker") {
		t.Error("missing emoji-picker class")
	}
	// Should contain some emojis
	if !strings.Contains(html, "📄") {
		t.Error("missing default document emoji")
	}
}
