//go:build !js || !wasm

package blocks_test

import (
	"strings"
	"testing"

	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/src/app/components/blocks"
	"github.com/Nu11ified/golem/src/app/models"
)

func TestRenderTextBlock(t *testing.T) {
	block := &models.Block{
		ID:      "b1",
		Type:    "text",
		Content: "Hello world",
		PageID:  "p1",
	}
	el := blocks.RenderBlock(block)
	if el == nil {
		t.Fatal("render should return element")
	}
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "Hello world") {
		t.Error("text block should contain content")
	}
	if !strings.Contains(html, "block-text") {
		t.Error("text block should have block-text class")
	}
}

func TestRenderH1Block(t *testing.T) {
	block := &models.Block{
		ID:      "b2",
		Type:    "h1",
		Content: "Heading One",
		PageID:  "p1",
	}
	el := blocks.RenderBlock(block)
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "Heading One") {
		t.Error("h1 block should contain content")
	}
	if !strings.Contains(html, "block-h1") {
		t.Error("h1 block should have block-h1 class")
	}
}

func TestRenderH2Block(t *testing.T) {
	block := &models.Block{
		ID:      "b3",
		Type:    "h2",
		Content: "Heading Two",
		PageID:  "p1",
	}
	el := blocks.RenderBlock(block)
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "Heading Two") {
		t.Error("h2 block should contain content")
	}
	if !strings.Contains(html, "block-h2") {
		t.Error("h2 block should have block-h2 class")
	}
}

func TestRenderH3Block(t *testing.T) {
	block := &models.Block{
		ID:      "b4",
		Type:    "h3",
		Content: "Heading Three",
		PageID:  "p1",
	}
	el := blocks.RenderBlock(block)
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "Heading Three") {
		t.Error("h3 block should contain content")
	}
	if !strings.Contains(html, "block-h3") {
		t.Error("h3 block should have block-h3 class")
	}
}

func TestRenderDividerBlock(t *testing.T) {
	block := &models.Block{
		ID:     "b5",
		Type:   "divider",
		PageID: "p1",
	}
	el := blocks.RenderBlock(block)
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "block-divider") {
		t.Error("divider block should have block-divider class")
	}
	if !strings.Contains(html, "hr") {
		t.Error("divider should contain an hr element")
	}
}

func TestRenderUnknownBlockType(t *testing.T) {
	block := &models.Block{
		ID:      "b6",
		Type:    "unknown_type",
		Content: "Some content",
		PageID:  "p1",
	}
	el := blocks.RenderBlock(block)
	if el == nil {
		t.Fatal("unknown block type should still render")
	}
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "Some content") {
		t.Error("unknown block should render content as text")
	}
}

func TestRenderBlockHasBlockClass(t *testing.T) {
	block := &models.Block{
		ID:      "b7",
		Type:    "text",
		Content: "Test",
		PageID:  "p1",
	}
	el := blocks.RenderBlock(block)
	html := dom.RenderToHTML(el)
	if !strings.Contains(html, "block") {
		t.Error("all blocks should have 'block' class")
	}
}

func TestRenderEmptyTextBlock(t *testing.T) {
	block := &models.Block{
		ID:     "b8",
		Type:   "text",
		PageID: "p1",
	}
	el := blocks.RenderBlock(block)
	if el == nil {
		t.Fatal("empty text block should still render")
	}
}
