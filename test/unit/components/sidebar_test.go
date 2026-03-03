//go:build !js || !wasm

package components_test

import (
	"strings"
	"testing"

	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/src/app/components"
	"github.com/Nu11ified/golem/src/app/models"
)

func TestSidebarStructure(t *testing.T) {
	ws := models.NewWorkspace()
	page := models.NewPage("Test Page", "")
	ws.Pages[page.ID] = page
	ws.RootPageIDs = append(ws.RootPageIDs, page.ID)

	sidebar := components.Sidebar(ws, page.ID)
	if sidebar == nil {
		t.Fatal("sidebar should not be nil")
	}

	html := dom.RenderToHTML(sidebar)
	if !strings.Contains(html, "sidebar") {
		t.Error("missing sidebar class")
	}
	if !strings.Contains(html, "sidebar-header") {
		t.Error("missing sidebar-header")
	}
}

func TestPageTreeRendersItems(t *testing.T) {
	ws := models.NewWorkspace()
	p1 := models.NewPage("Page One", "")
	p2 := models.NewPage("Page Two", "")
	ws.Pages[p1.ID] = p1
	ws.Pages[p2.ID] = p2
	ws.RootPageIDs = []string{p1.ID, p2.ID}

	tree := components.PageTree(ws, "")
	if tree == nil {
		t.Fatal("page tree should not be nil")
	}

	html := dom.RenderToHTML(tree)
	if !strings.Contains(html, "Page One") {
		t.Error("missing Page One")
	}
	if !strings.Contains(html, "Page Two") {
		t.Error("missing Page Two")
	}
}

func TestPageTreeActiveItem(t *testing.T) {
	ws := models.NewWorkspace()
	page := models.NewPage("Active", "")
	ws.Pages[page.ID] = page
	ws.RootPageIDs = []string{page.ID}

	tree := components.PageTree(ws, page.ID)
	html := dom.RenderToHTML(tree)
	if !strings.Contains(html, "page-tree-item-active") {
		t.Error("active page should have active class")
	}
}

func TestPageTreeNestedItems(t *testing.T) {
	ws := models.NewWorkspace()
	parent := models.NewPage("Parent", "")
	child := models.NewPage("Child", parent.ID)
	parent.ChildIDs = []string{child.ID}
	ws.Pages[parent.ID] = parent
	ws.Pages[child.ID] = child
	ws.RootPageIDs = []string{parent.ID}

	tree := components.PageTree(ws, "")
	html := dom.RenderToHTML(tree)
	if !strings.Contains(html, "Parent") {
		t.Error("missing parent")
	}
	if !strings.Contains(html, "Child") {
		t.Error("missing child")
	}
}

func TestPageTreeEmpty(t *testing.T) {
	ws := models.NewWorkspace()
	tree := components.PageTree(ws, "")
	if tree == nil {
		t.Fatal("empty page tree should still return an element")
	}
}

func TestSidebarHasNewPageButton(t *testing.T) {
	ws := models.NewWorkspace()
	sidebar := components.Sidebar(ws, "")
	html := dom.RenderToHTML(sidebar)
	if !strings.Contains(html, "new-page-btn") {
		t.Error("missing new page button")
	}
}

func TestSidebarHasHamburgerButton(t *testing.T) {
	ws := models.NewWorkspace()
	sidebar := components.Sidebar(ws, "")
	html := dom.RenderToHTML(sidebar)
	if !strings.Contains(html, "hamburger-btn") {
		t.Error("missing hamburger button for mobile")
	}
}
