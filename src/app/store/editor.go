//go:build js && wasm

package store

import (
	"github.com/Nu11ified/golem/src/app/models"
	"github.com/Nu11ified/golem/state"
)

// Action type constants for editor operations
const (
	LoadPage        = "LOAD_PAGE"
	AddBlock        = "ADD_BLOCK"
	DeleteBlock     = "DELETE_BLOCK"
	UpdateBlock     = "UPDATE_BLOCK"
	MoveBlock       = "MOVE_BLOCK"
	ToggleBlock     = "TOGGLE_BLOCK"
	ChangeBlockType = "CHANGE_BLOCK_TYPE"
)

// EditorReducer handles editor state transitions
func EditorReducer(s interface{}, action state.Action) interface{} {
	es, ok := s.(*models.EditorState)
	if !ok {
		return s
	}

	switch action.Type {
	case LoadPage:
		return handleLoadPage(es, action.Payload)
	case AddBlock:
		return handleAddBlock(es, action.Payload)
	case DeleteBlock:
		return handleDeleteBlock(es, action.Payload)
	case UpdateBlock:
		return handleUpdateBlock(es, action.Payload)
	case ChangeBlockType:
		return handleChangeBlockType(es, action.Payload)
	case MoveBlock:
		return handleMoveBlock(es, action.Payload)
	case ToggleBlock:
		return handleToggleBlock(es, action.Payload)
	default:
		return es
	}
}

func handleLoadPage(es *models.EditorState, payload interface{}) *models.EditorState {
	data := payload.(map[string]interface{})
	es.ActivePageID = data["pageId"].(string)
	es.Blocks = data["blocks"].(map[string]*models.Block)
	es.BlockOrder = data["blockOrder"].([]string)
	es.FocusBlockID = ""
	return es
}

func handleAddBlock(es *models.EditorState, payload interface{}) *models.EditorState {
	data := payload.(map[string]interface{})
	blockType := data["blockType"].(string)
	afterBlockID := data["afterBlockID"].(string)
	pageID := data["pageId"].(string)

	block := models.NewBlock(blockType, pageID)
	es.Blocks[block.ID] = block

	if afterBlockID == "" {
		// Add at end
		es.BlockOrder = append(es.BlockOrder, block.ID)
	} else {
		// Insert after the specified block
		for i, id := range es.BlockOrder {
			if id == afterBlockID {
				newOrder := make([]string, 0, len(es.BlockOrder)+1)
				newOrder = append(newOrder, es.BlockOrder[:i+1]...)
				newOrder = append(newOrder, block.ID)
				newOrder = append(newOrder, es.BlockOrder[i+1:]...)
				es.BlockOrder = newOrder
				break
			}
		}
	}

	es.FocusBlockID = block.ID
	return es
}

func handleDeleteBlock(es *models.EditorState, payload interface{}) *models.EditorState {
	blockID := payload.(string)
	delete(es.Blocks, blockID)

	for i, id := range es.BlockOrder {
		if id == blockID {
			es.BlockOrder = append(es.BlockOrder[:i], es.BlockOrder[i+1:]...)
			// Focus previous block
			if i > 0 {
				es.FocusBlockID = es.BlockOrder[i-1]
			} else if len(es.BlockOrder) > 0 {
				es.FocusBlockID = es.BlockOrder[0]
			} else {
				es.FocusBlockID = ""
			}
			break
		}
	}

	return es
}

func handleUpdateBlock(es *models.EditorState, payload interface{}) *models.EditorState {
	data := payload.(map[string]interface{})
	id := data["id"].(string)
	content := data["content"].(string)

	if block, ok := es.Blocks[id]; ok {
		block.Content = content
	}
	return es
}

func handleChangeBlockType(es *models.EditorState, payload interface{}) *models.EditorState {
	data := payload.(map[string]interface{})
	id := data["id"].(string)
	newType := data["newType"].(string)

	if block, ok := es.Blocks[id]; ok {
		block.Type = newType
	}
	return es
}

func handleMoveBlock(es *models.EditorState, payload interface{}) *models.EditorState {
	data := payload.(map[string]interface{})
	id := data["id"].(string)
	newIndex := data["newIndex"].(int)

	// Remove from current position
	var oldIndex int
	for i, bid := range es.BlockOrder {
		if bid == id {
			oldIndex = i
			break
		}
	}
	es.BlockOrder = append(es.BlockOrder[:oldIndex], es.BlockOrder[oldIndex+1:]...)

	// Insert at new position
	if newIndex >= len(es.BlockOrder) {
		es.BlockOrder = append(es.BlockOrder, id)
	} else {
		newOrder := make([]string, 0, len(es.BlockOrder)+1)
		newOrder = append(newOrder, es.BlockOrder[:newIndex]...)
		newOrder = append(newOrder, id)
		newOrder = append(newOrder, es.BlockOrder[newIndex:]...)
		es.BlockOrder = newOrder
	}

	return es
}

func handleToggleBlock(es *models.EditorState, payload interface{}) *models.EditorState {
	blockID := payload.(string)
	if block, ok := es.Blocks[blockID]; ok {
		collapsed, _ := block.Props["collapsed"].(bool)
		block.Props["collapsed"] = !collapsed
	}
	return es
}
