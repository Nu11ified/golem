//go:build !js || !wasm

package store

import (
	"time"

	"github.com/Nu11ified/golem/src/app/models"
	"github.com/Nu11ified/golem/state"
)

// Action type constants for workspace operations
const (
	CreatePage   = "CREATE_PAGE"
	DeletePage   = "DELETE_PAGE"
	RenamePage    = "RENAME_PAGE"
	MovePage     = "MOVE_PAGE"
	SetPageIcon  = "SET_PAGE_ICON"
	ReorderPages = "REORDER_PAGES"
)

// WorkspaceReducer handles workspace state transitions
func WorkspaceReducer(s interface{}, action state.Action) interface{} {
	ws, ok := s.(*models.Workspace)
	if !ok {
		return s
	}

	switch action.Type {
	case CreatePage:
		return handleCreatePage(ws, action.Payload)
	case DeletePage:
		return handleDeletePage(ws, action.Payload)
	case RenamePage:
		return handleRenamePage(ws, action.Payload)
	case SetPageIcon:
		return handleSetPageIcon(ws, action.Payload)
	case MovePage:
		return handleMovePage(ws, action.Payload)
	default:
		return ws
	}
}

func handleCreatePage(ws *models.Workspace, payload interface{}) *models.Workspace {
	data := payload.(map[string]interface{})
	title := data["title"].(string)
	parentID := data["parentId"].(string)

	page := models.NewPage(title, parentID)
	ws.Pages[page.ID] = page

	if parentID == "" {
		ws.RootPageIDs = append(ws.RootPageIDs, page.ID)
	} else if parent, ok := ws.Pages[parentID]; ok {
		parent.ChildIDs = append(parent.ChildIDs, page.ID)
		parent.UpdatedAt = time.Now().UnixMilli()
	}

	return ws
}

func handleDeletePage(ws *models.Workspace, payload interface{}) *models.Workspace {
	pageID := payload.(string)
	page, ok := ws.Pages[pageID]
	if !ok {
		return ws
	}

	// Remove from parent's children or from root
	if page.ParentID == "" {
		for i, id := range ws.RootPageIDs {
			if id == pageID {
				ws.RootPageIDs = append(ws.RootPageIDs[:i], ws.RootPageIDs[i+1:]...)
				break
			}
		}
	} else if parent, ok := ws.Pages[page.ParentID]; ok {
		for i, id := range parent.ChildIDs {
			if id == pageID {
				parent.ChildIDs = append(parent.ChildIDs[:i], parent.ChildIDs[i+1:]...)
				break
			}
		}
	}

	// Recursively delete children
	var deleteChildren func(id string)
	deleteChildren = func(id string) {
		if p, ok := ws.Pages[id]; ok {
			for _, childID := range p.ChildIDs {
				deleteChildren(childID)
			}
			delete(ws.Pages, id)
		}
	}
	deleteChildren(pageID)

	return ws
}

func handleRenamePage(ws *models.Workspace, payload interface{}) *models.Workspace {
	data := payload.(map[string]interface{})
	id := data["id"].(string)
	title := data["title"].(string)
	if page, ok := ws.Pages[id]; ok {
		page.Title = title
		page.UpdatedAt = time.Now().UnixMilli()
	}
	return ws
}

func handleSetPageIcon(ws *models.Workspace, payload interface{}) *models.Workspace {
	data := payload.(map[string]interface{})
	id := data["id"].(string)
	icon := data["icon"].(string)
	if page, ok := ws.Pages[id]; ok {
		page.Icon = icon
		page.UpdatedAt = time.Now().UnixMilli()
	}
	return ws
}

func handleMovePage(ws *models.Workspace, payload interface{}) *models.Workspace {
	data := payload.(map[string]interface{})
	id := data["id"].(string)
	newParentID := data["newParentId"].(string)

	page, ok := ws.Pages[id]
	if !ok {
		return ws
	}

	// Remove from current parent
	if page.ParentID == "" {
		for i, pid := range ws.RootPageIDs {
			if pid == id {
				ws.RootPageIDs = append(ws.RootPageIDs[:i], ws.RootPageIDs[i+1:]...)
				break
			}
		}
	} else if parent, ok := ws.Pages[page.ParentID]; ok {
		for i, pid := range parent.ChildIDs {
			if pid == id {
				parent.ChildIDs = append(parent.ChildIDs[:i], parent.ChildIDs[i+1:]...)
				break
			}
		}
	}

	// Add to new parent
	page.ParentID = newParentID
	if newParentID == "" {
		ws.RootPageIDs = append(ws.RootPageIDs, id)
	} else if newParent, ok := ws.Pages[newParentID]; ok {
		newParent.ChildIDs = append(newParent.ChildIDs, id)
	}

	page.UpdatedAt = time.Now().UnixMilli()
	return ws
}
