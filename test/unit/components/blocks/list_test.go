//go:build !js || !wasm

package blocks_test

import (
	"strings"
	"testing"

	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/src/app/components/blocks"
	"github.com/Nu11ified/golem/src/app/models"
)

func TestBulletListBlock(t *testing.T) {
	block := &models.Block{
		ID:      "b1",
		Type:    "bullet",
		Content: "Bullet item",
		PageID:  "p1",
	}
	el := blocks.RenderBlock(block)
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "block-bullet") {
		t.Error("missing block-bullet class")
	}
	if !strings.Contains(html, "Bullet item") {
		t.Error("missing content")
	}
	// Should have a bullet marker
	if !strings.Contains(html, "block-bullet-marker") {
		t.Error("missing bullet marker")
	}
}

func TestNumberedListBlock(t *testing.T) {
	block := &models.Block{
		ID:      "b2",
		Type:    "numbered",
		Content: "Numbered item",
		PageID:  "p1",
	}
	el := blocks.RenderBlock(block)
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "block-numbered") {
		t.Error("missing block-numbered class")
	}
	if !strings.Contains(html, "Numbered item") {
		t.Error("missing content")
	}
	if !strings.Contains(html, "block-number") {
		t.Error("missing number marker")
	}
}

func TestToggleBlockCollapsed(t *testing.T) {
	block := &models.Block{
		ID:       "b3",
		Type:     "toggle",
		Content:  "Toggle content",
		Children: []string{},
		Props:    map[string]interface{}{},
		PageID:   "p1",
	}
	el := blocks.RenderBlock(block)
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "block-toggle") {
		t.Error("missing block-toggle class")
	}
	if !strings.Contains(html, "Toggle content") {
		t.Error("missing content")
	}
	if !strings.Contains(html, "toggle-indicator") {
		t.Error("missing toggle indicator")
	}
}

func TestToggleBlockExpanded(t *testing.T) {
	childBlock := &models.Block{
		ID:      "child-1",
		Type:    "text",
		Content: "Child content",
		PageID:  "p1",
	}
	block := &models.Block{
		ID:       "b4",
		Type:     "toggle",
		Content:  "Parent toggle",
		Children: []string{childBlock.ID},
		Props:    map[string]interface{}{"collapsed": false},
		PageID:   "p1",
	}

	// Use the version with child blocks map
	el := blocks.RenderToggleWithChildren(block, map[string]*models.Block{
		childBlock.ID: childBlock,
	})
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "expanded") {
		t.Error("expanded toggle should have expanded class on indicator")
	}
	if !strings.Contains(html, "toggle-children") {
		t.Error("expanded toggle should have children container")
	}
	if !strings.Contains(html, "Child content") {
		t.Error("expanded toggle should render child content")
	}
}

func TestCodeBlock(t *testing.T) {
	block := &models.Block{
		ID:      "b5",
		Type:    "code",
		Content: "fmt.Println(\"hello\")",
		Props:   map[string]interface{}{"language": "go"},
		PageID:  "p1",
	}
	el := blocks.RenderBlock(block)
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "block-code") {
		t.Error("missing block-code class")
	}
	if !strings.Contains(html, "fmt.Println") {
		t.Error("missing code content")
	}
}

func TestCodeBlockWithLanguage(t *testing.T) {
	block := &models.Block{
		ID:      "b6",
		Type:    "code",
		Content: "console.log('hi')",
		Props:   map[string]interface{}{"language": "javascript"},
		PageID:  "p1",
	}
	el := blocks.RenderBlock(block)
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "javascript") {
		t.Error("missing language label")
	}
}
