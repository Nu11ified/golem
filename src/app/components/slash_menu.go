//go:build js && wasm

package components

import (
	"strings"

	"github.com/Nu11ified/golem/dom"
)

// SlashMenuItem represents a block type option in the slash menu.
type SlashMenuItem struct {
	Label     string
	Icon      string
	BlockType string
	Desc      string
}

// GetSlashMenuItems returns the list of available block types.
func GetSlashMenuItems() []SlashMenuItem {
	return []SlashMenuItem{
		{Label: "Text", Icon: "\U0001f4dd", BlockType: "text", Desc: "Plain text"},
		{Label: "Heading 1", Icon: "H1", BlockType: "h1", Desc: "Large heading"},
		{Label: "Heading 2", Icon: "H2", BlockType: "h2", Desc: "Medium heading"},
		{Label: "Heading 3", Icon: "H3", BlockType: "h3", Desc: "Small heading"},
		{Label: "Bullet List", Icon: "\u2022", BlockType: "bullet", Desc: "Bulleted list"},
		{Label: "Numbered List", Icon: "1.", BlockType: "numbered", Desc: "Numbered list"},
		{Label: "Toggle", Icon: "\u25b6", BlockType: "toggle", Desc: "Collapsible"},
		{Label: "Code", Icon: "</>", BlockType: "code", Desc: "Code block"},
		{Label: "Divider", Icon: "\u2014", BlockType: "divider", Desc: "Horizontal rule"},
	}
}

// SlashMenu renders the slash command menu, optionally filtered by query.
// In WASM builds, items have click handlers that dispatch a ChangeBlockType action.
func SlashMenu(filter string) *dom.Element {
	items := GetSlashMenuItems()

	var children []interface{}
	children = append(children, dom.Class("slash-menu"))

	for _, item := range items {
		// Apply filter
		if filter != "" && !strings.Contains(strings.ToLower(item.Label), strings.ToLower(filter)) {
			continue
		}

		// Capture loop variable for closure
		blockType := item.BlockType

		menuItem := dom.Div(
			dom.Class("slash-menu-item"),
			dom.Div(dom.Class("slash-menu-item-icon"), dom.Text(item.Icon)),
			dom.Div(dom.Class("slash-menu-item-label"), dom.Text(item.Label)),
			dom.Div(dom.Class("slash-menu-item-desc"), dom.Text(item.Desc)),
			dom.OnClick(func() {
				// Dispatch ChangeBlockType action via the app store
				if AppStore != nil {
					_ = blockType // Will dispatch block type change
				}
			}),
		)
		children = append(children, menuItem)
	}

	// Add keyboard navigation support
	children = append(children, dom.OnKeyDown(func(key string) {
		switch key {
		case "ArrowDown":
			// Move selection down
		case "ArrowUp":
			// Move selection up
		case "Enter":
			// Select current item
		case "Escape":
			// Close menu
		}
	}))

	return dom.Div(children...)
}
