//go:build !js || !wasm

package store_test

import (
	"testing"

	"github.com/Nu11ified/golem/src/app/models"
	"github.com/Nu11ified/golem/src/app/store"
	golemstate "github.com/Nu11ified/golem/state"
)

func TestLoadPage(t *testing.T) {
	es := models.NewEditorState()
	block := models.NewBlock("text", "page-1")

	action := golemstate.Action{
		Type: store.LoadPage,
		Payload: map[string]interface{}{
			"pageId":     "page-1",
			"blocks":     map[string]*models.Block{block.ID: block},
			"blockOrder": []string{block.ID},
		},
	}
	result := store.EditorReducer(es, action)
	es2 := result.(*models.EditorState)

	if es2.ActivePageID != "page-1" {
		t.Errorf("expected activePageId 'page-1', got %q", es2.ActivePageID)
	}
	if len(es2.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(es2.Blocks))
	}
	if len(es2.BlockOrder) != 1 {
		t.Fatalf("expected 1 block in order, got %d", len(es2.BlockOrder))
	}
}

func TestAddBlock(t *testing.T) {
	es := models.NewEditorState()
	es.ActivePageID = "page-1"

	// Add first block
	action := golemstate.Action{
		Type: store.AddBlock,
		Payload: map[string]interface{}{
			"blockType":    "text",
			"afterBlockID": "",
			"pageId":       "page-1",
		},
	}
	result := store.EditorReducer(es, action)
	es2 := result.(*models.EditorState)

	if len(es2.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(es2.Blocks))
	}
	if len(es2.BlockOrder) != 1 {
		t.Fatalf("expected 1 in order, got %d", len(es2.BlockOrder))
	}
	firstBlockID := es2.BlockOrder[0]

	// Add second block after first
	action2 := golemstate.Action{
		Type: store.AddBlock,
		Payload: map[string]interface{}{
			"blockType":    "text",
			"afterBlockID": firstBlockID,
			"pageId":       "page-1",
		},
	}
	result2 := store.EditorReducer(es2, action2)
	es3 := result2.(*models.EditorState)

	if len(es3.BlockOrder) != 2 {
		t.Fatalf("expected 2 in order, got %d", len(es3.BlockOrder))
	}
	if es3.BlockOrder[0] != firstBlockID {
		t.Error("first block should remain at index 0")
	}
}

func TestDeleteBlock(t *testing.T) {
	es := models.NewEditorState()
	es.ActivePageID = "page-1"
	block := models.NewBlock("text", "page-1")
	es.Blocks[block.ID] = block
	es.BlockOrder = []string{block.ID}

	result := store.EditorReducer(es, golemstate.Action{
		Type:    store.DeleteBlock,
		Payload: block.ID,
	})
	es2 := result.(*models.EditorState)

	if len(es2.Blocks) != 0 {
		t.Errorf("expected 0 blocks, got %d", len(es2.Blocks))
	}
	if len(es2.BlockOrder) != 0 {
		t.Errorf("expected 0 in order, got %d", len(es2.BlockOrder))
	}
}

func TestUpdateBlock(t *testing.T) {
	es := models.NewEditorState()
	block := models.NewBlock("text", "page-1")
	block.Content = "old"
	es.Blocks[block.ID] = block
	es.BlockOrder = []string{block.ID}

	result := store.EditorReducer(es, golemstate.Action{
		Type: store.UpdateBlock,
		Payload: map[string]interface{}{
			"id":      block.ID,
			"content": "new content",
		},
	})
	es2 := result.(*models.EditorState)

	if es2.Blocks[block.ID].Content != "new content" {
		t.Errorf("expected 'new content', got %q", es2.Blocks[block.ID].Content)
	}
}

func TestChangeBlockType(t *testing.T) {
	es := models.NewEditorState()
	block := models.NewBlock("text", "page-1")
	es.Blocks[block.ID] = block
	es.BlockOrder = []string{block.ID}

	result := store.EditorReducer(es, golemstate.Action{
		Type: store.ChangeBlockType,
		Payload: map[string]interface{}{
			"id":      block.ID,
			"newType": "h1",
		},
	})
	es2 := result.(*models.EditorState)

	if es2.Blocks[block.ID].Type != "h1" {
		t.Errorf("expected type 'h1', got %q", es2.Blocks[block.ID].Type)
	}
}

func TestMoveBlock(t *testing.T) {
	es := models.NewEditorState()
	b1 := models.NewBlock("text", "page-1")
	b2 := models.NewBlock("text", "page-1")
	b3 := models.NewBlock("text", "page-1")
	es.Blocks[b1.ID] = b1
	es.Blocks[b2.ID] = b2
	es.Blocks[b3.ID] = b3
	es.BlockOrder = []string{b1.ID, b2.ID, b3.ID}

	// Move b3 to index 0
	result := store.EditorReducer(es, golemstate.Action{
		Type: store.MoveBlock,
		Payload: map[string]interface{}{
			"id":       b3.ID,
			"newIndex": 0,
		},
	})
	es2 := result.(*models.EditorState)

	if es2.BlockOrder[0] != b3.ID {
		t.Errorf("expected b3 at index 0, got %s", es2.BlockOrder[0])
	}
	if len(es2.BlockOrder) != 3 {
		t.Errorf("expected 3 blocks, got %d", len(es2.BlockOrder))
	}
}

func TestToggleBlock(t *testing.T) {
	es := models.NewEditorState()
	block := models.NewBlock("toggle", "page-1")
	es.Blocks[block.ID] = block
	es.BlockOrder = []string{block.ID}

	// Toggle once -- should set collapsed to true
	result := store.EditorReducer(es, golemstate.Action{
		Type:    store.ToggleBlock,
		Payload: block.ID,
	})
	es2 := result.(*models.EditorState)

	collapsed, ok := es2.Blocks[block.ID].Props["collapsed"].(bool)
	if !ok || !collapsed {
		t.Error("expected collapsed to be true")
	}

	// Toggle again -- should set collapsed to false
	result2 := store.EditorReducer(es2, golemstate.Action{
		Type:    store.ToggleBlock,
		Payload: block.ID,
	})
	es3 := result2.(*models.EditorState)

	collapsed2, ok := es3.Blocks[block.ID].Props["collapsed"].(bool)
	if !ok || collapsed2 {
		t.Error("expected collapsed to be false")
	}
}

func TestEditorUnknownAction(t *testing.T) {
	es := models.NewEditorState()
	result := store.EditorReducer(es, golemstate.Action{Type: "UNKNOWN"})
	if result != es {
		t.Error("unknown action should return state unchanged")
	}
}
