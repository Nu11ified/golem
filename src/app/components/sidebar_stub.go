//go:build !js || !wasm

package components

import (
	"fmt"

	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/src/app/models"
)

// Sidebar creates the sidebar component with hamburger toggle, header, and page tree.
func Sidebar(ws *models.Workspace, activePageID string) *dom.Element {
	// Hamburger button for mobile
	hamburger := dom.Button(
		dom.Class("hamburger-btn"),
		dom.Text("\u2630"),
	)

	// Sidebar overlay for mobile backdrop
	overlay := dom.Div(dom.Class("sidebar-overlay"))

	// Sidebar header with workspace title and new page button
	header := dom.Div(
		dom.Class("sidebar-header"),
		dom.Span(dom.Class("sidebar-title"), dom.Text("My Workspace")),
		dom.Button(dom.Class("new-page-btn"), dom.Text("+")),
	)

	// Page tree
	pageTree := PageTree(ws, activePageID)

	// Sidebar container
	sidebar := dom.Div(
		dom.Class("sidebar"),
		header,
		pageTree,
	)

	// Wrapper containing hamburger, overlay, and sidebar
	return dom.Div(
		dom.Class("sidebar-wrapper"),
		hamburger,
		overlay,
		sidebar,
	)
}

// PageTree renders the tree of pages in the workspace.
func PageTree(ws *models.Workspace, activePageID string) *dom.Element {
	tree := dom.Div(dom.Class("page-tree"))

	for _, pageID := range ws.RootPageIDs {
		if page, ok := ws.Pages[pageID]; ok {
			item := PageTreeItem(page, 0, activePageID, ws)
			tree.AddChild(item)
		}
	}

	return tree
}

// PageTreeItem renders a single page item in the tree, with recursive children.
func PageTreeItem(page *models.Page, depth int, activePageID string, ws *models.Workspace) *dom.Element {
	// Determine CSS class
	className := "page-tree-item"
	if page.ID == activePageID {
		className = "page-tree-item page-tree-item-active"
	}

	// Calculate padding based on depth
	paddingLeft := fmt.Sprintf("padding-left: %dpx;", 14+depth*16)

	// Item row with icon and title
	row := dom.Div(
		dom.Class(className),
		dom.Style(map[string]string{"padding-left": fmt.Sprintf("%dpx", 14+depth*16)}),
		dom.Span(dom.Class("page-tree-item-icon"), dom.Text(page.Icon)),
		dom.Span(dom.Class("page-tree-item-title"), dom.Text(page.Title)),
	)

	// Container for item + children
	container := dom.Div(dom.Class("page-tree-item-container"))
	container.AddChild(row)

	// Recursively render children
	for _, childID := range page.ChildIDs {
		if child, ok := ws.Pages[childID]; ok {
			childItem := PageTreeItem(child, depth+1, activePageID, ws)
			container.AddChild(childItem)
		}
	}

	_ = paddingLeft // suppress unused warning
	return container
}
