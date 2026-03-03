//go:build !js || !wasm

package blocks

import (
	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/src/app/models"
)

// isToggleExpanded checks whether a toggle block is expanded.
// A toggle is expanded when its "collapsed" prop is explicitly set to false.
func isToggleExpanded(block *models.Block) bool {
	if block.Props == nil {
		return false
	}
	collapsed, ok := block.Props["collapsed"]
	if !ok {
		return false
	}
	if b, isBool := collapsed.(bool); isBool {
		return !b
	}
	return false
}

// ToggleBlock renders a collapsible toggle block.
// When collapsed (default), shows a right arrow. When expanded, shows a down arrow.
func ToggleBlock(block *models.Block) *dom.Element {
	expanded := isToggleExpanded(block)

	indicator := "\u25b6"
	indicatorClass := "toggle-indicator"
	if expanded {
		indicator = "\u25bc"
		indicatorClass = "toggle-indicator expanded"
	}

	return dom.Div(
		dom.Class("block block-toggle"),
		dom.Div(dom.Class(indicatorClass), dom.Text(indicator)),
		dom.Div(dom.Class("block-content"), dom.Text(block.Content)),
	)
}

// RenderToggleWithChildren renders a toggle block with its child blocks.
// When expanded, child blocks are rendered inside a toggle-children container.
func RenderToggleWithChildren(block *models.Block, childBlocks map[string]*models.Block) *dom.Element {
	expanded := isToggleExpanded(block)

	indicator := "\u25b6"
	indicatorClass := "toggle-indicator"
	if expanded {
		indicator = "\u25bc"
		indicatorClass = "toggle-indicator expanded"
	}

	wrapper := dom.Div(
		dom.Class("block block-toggle"),
		dom.Div(dom.Class(indicatorClass), dom.Text(indicator)),
		dom.Div(dom.Class("block-content"), dom.Text(block.Content)),
	)

	if expanded && len(block.Children) > 0 {
		childrenContainer := dom.Div(dom.Class("toggle-children"))
		for _, childID := range block.Children {
			if child, ok := childBlocks[childID]; ok {
				childrenContainer.AddChild(RenderBlock(child))
			}
		}
		wrapper.AddChild(childrenContainer)
	}

	return wrapper
}
