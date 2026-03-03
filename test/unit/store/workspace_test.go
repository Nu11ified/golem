//go:build !js || !wasm

package store_test

import (
	"testing"

	"github.com/Nu11ified/golem/src/app/models"
	"github.com/Nu11ified/golem/src/app/store"
	golemstate "github.com/Nu11ified/golem/state"
)

func TestCreatePage(t *testing.T) {
	ws := models.NewWorkspace()
	action := golemstate.Action{
		Type:    store.CreatePage,
		Payload: map[string]interface{}{"title": "New Page", "parentId": ""},
	}
	result := store.WorkspaceReducer(ws, action)
	ws2 := result.(*models.Workspace)
	if len(ws2.RootPageIDs) != 1 {
		t.Fatalf("expected 1 root page, got %d", len(ws2.RootPageIDs))
	}
	page := ws2.Pages[ws2.RootPageIDs[0]]
	if page.Title != "New Page" {
		t.Errorf("expected title 'New Page', got %q", page.Title)
	}
}

func TestDeletePage(t *testing.T) {
	ws := models.NewWorkspace()
	ws = store.WorkspaceReducer(ws, golemstate.Action{
		Type:    store.CreatePage,
		Payload: map[string]interface{}{"title": "To Delete", "parentId": ""},
	}).(*models.Workspace)
	pageID := ws.RootPageIDs[0]

	ws = store.WorkspaceReducer(ws, golemstate.Action{
		Type:    store.DeletePage,
		Payload: pageID,
	}).(*models.Workspace)

	if len(ws.RootPageIDs) != 0 {
		t.Errorf("expected 0 root pages, got %d", len(ws.RootPageIDs))
	}
	if _, exists := ws.Pages[pageID]; exists {
		t.Error("page should be deleted from map")
	}
}

func TestRenamePage(t *testing.T) {
	ws := models.NewWorkspace()
	ws = store.WorkspaceReducer(ws, golemstate.Action{
		Type:    store.CreatePage,
		Payload: map[string]interface{}{"title": "Old", "parentId": ""},
	}).(*models.Workspace)
	pageID := ws.RootPageIDs[0]

	ws = store.WorkspaceReducer(ws, golemstate.Action{
		Type:    store.RenamePage,
		Payload: map[string]interface{}{"id": pageID, "title": "New"},
	}).(*models.Workspace)

	if ws.Pages[pageID].Title != "New" {
		t.Errorf("expected 'New', got %q", ws.Pages[pageID].Title)
	}
}

func TestCreateChildPage(t *testing.T) {
	ws := models.NewWorkspace()
	ws = store.WorkspaceReducer(ws, golemstate.Action{
		Type:    store.CreatePage,
		Payload: map[string]interface{}{"title": "Parent", "parentId": ""},
	}).(*models.Workspace)
	parentID := ws.RootPageIDs[0]

	ws = store.WorkspaceReducer(ws, golemstate.Action{
		Type:    store.CreatePage,
		Payload: map[string]interface{}{"title": "Child", "parentId": parentID},
	}).(*models.Workspace)

	if len(ws.RootPageIDs) != 1 {
		t.Fatalf("child should not be in root pages")
	}
	parent := ws.Pages[parentID]
	if len(parent.ChildIDs) != 1 {
		t.Fatalf("expected 1 child, got %d", len(parent.ChildIDs))
	}
	child := ws.Pages[parent.ChildIDs[0]]
	if child.Title != "Child" || child.ParentID != parentID {
		t.Errorf("unexpected child: %+v", child)
	}
}

func TestSetPageIcon(t *testing.T) {
	ws := models.NewWorkspace()
	ws = store.WorkspaceReducer(ws, golemstate.Action{
		Type:    store.CreatePage,
		Payload: map[string]interface{}{"title": "Test", "parentId": ""},
	}).(*models.Workspace)
	pageID := ws.RootPageIDs[0]

	ws = store.WorkspaceReducer(ws, golemstate.Action{
		Type:    store.SetPageIcon,
		Payload: map[string]interface{}{"id": pageID, "icon": "\xf0\x9f\x9a\x80"},
	}).(*models.Workspace)

	if ws.Pages[pageID].Icon != "\xf0\x9f\x9a\x80" {
		t.Errorf("expected icon '\xf0\x9f\x9a\x80', got %q", ws.Pages[pageID].Icon)
	}
}

func TestDeleteChildPage(t *testing.T) {
	ws := models.NewWorkspace()
	ws = store.WorkspaceReducer(ws, golemstate.Action{
		Type:    store.CreatePage,
		Payload: map[string]interface{}{"title": "Parent", "parentId": ""},
	}).(*models.Workspace)
	parentID := ws.RootPageIDs[0]

	ws = store.WorkspaceReducer(ws, golemstate.Action{
		Type:    store.CreatePage,
		Payload: map[string]interface{}{"title": "Child", "parentId": parentID},
	}).(*models.Workspace)
	childID := ws.Pages[parentID].ChildIDs[0]

	ws = store.WorkspaceReducer(ws, golemstate.Action{
		Type:    store.DeletePage,
		Payload: childID,
	}).(*models.Workspace)

	if len(ws.Pages[parentID].ChildIDs) != 0 {
		t.Errorf("expected 0 children, got %d", len(ws.Pages[parentID].ChildIDs))
	}
	if _, exists := ws.Pages[childID]; exists {
		t.Error("child page should be deleted")
	}
}

func TestUnknownActionReturnsState(t *testing.T) {
	ws := models.NewWorkspace()
	result := store.WorkspaceReducer(ws, golemstate.Action{
		Type: "UNKNOWN",
	})
	if result != ws {
		t.Error("unknown action should return state unchanged")
	}
}
