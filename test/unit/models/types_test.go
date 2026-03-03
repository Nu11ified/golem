//go:build !js || !wasm

package models_test

import (
	"encoding/json"
	"testing"

	"github.com/Nu11ified/golem/src/app/models"
)

func TestPageJSON(t *testing.T) {
	p := &models.Page{
		ID:       "page-1",
		Title:    "Test Page",
		ParentID: "",
		ChildIDs: []string{"page-2"},
		Icon:     "\xf0\x9f\x93\x84",
	}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var p2 models.Page
	if err := json.Unmarshal(data, &p2); err != nil {
		t.Fatal(err)
	}
	if p2.ID != "page-1" || p2.Title != "Test Page" || p2.Icon != "\xf0\x9f\x93\x84" {
		t.Errorf("round-trip mismatch: %+v", p2)
	}
	if len(p2.ChildIDs) != 1 || p2.ChildIDs[0] != "page-2" {
		t.Errorf("childIDs mismatch: %v", p2.ChildIDs)
	}
}

func TestBlockJSON(t *testing.T) {
	b := &models.Block{
		ID:      "block-1",
		Type:    "text",
		Content: "Hello world",
		PageID:  "page-1",
	}
	data, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	var b2 models.Block
	if err := json.Unmarshal(data, &b2); err != nil {
		t.Fatal(err)
	}
	if b2.Type != "text" || b2.Content != "Hello world" {
		t.Errorf("round-trip mismatch: %+v", b2)
	}
}

func TestNewWorkspace(t *testing.T) {
	ws := models.NewWorkspace()
	if ws.Pages == nil {
		t.Fatal("Pages map should be initialized")
	}
	if ws.RootPageIDs == nil {
		t.Fatal("RootPageIDs should be initialized")
	}
}

func TestNewEditorState(t *testing.T) {
	es := models.NewEditorState()
	if es.Blocks == nil {
		t.Fatal("Blocks map should be initialized")
	}
	if es.BlockOrder == nil {
		t.Fatal("BlockOrder should be initialized")
	}
}

func TestNewPage(t *testing.T) {
	p := models.NewPage("Test", "")
	if p.ID == "" {
		t.Fatal("ID should be generated")
	}
	if p.Title != "Test" {
		t.Errorf("expected title 'Test', got %q", p.Title)
	}
	if p.CreatedAt == 0 {
		t.Error("CreatedAt should be set")
	}
}

func TestNewBlock(t *testing.T) {
	b := models.NewBlock("text", "page-1")
	if b.ID == "" {
		t.Fatal("ID should be generated")
	}
	if b.Type != "text" || b.PageID != "page-1" {
		t.Errorf("unexpected block: %+v", b)
	}
}
