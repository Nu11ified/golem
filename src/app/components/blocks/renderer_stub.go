//go:build !js || !wasm

package blocks

import (
	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/src/app/models"
)

// RenderBlock dispatches to the correct block renderer based on block type.
func RenderBlock(block *models.Block) *dom.Element {
	switch block.Type {
	case "text":
		return TextBlock(block)
	case "h1", "h2", "h3":
		return HeadingBlock(block)
	case "divider":
		return DividerBlock(block)
	case "bullet":
		return BulletListBlock(block)
	case "numbered":
		return NumberedListBlock(block)
	case "toggle":
		return ToggleBlock(block)
	case "code":
		return CodeBlock(block)
	default:
		// Unknown type — render as text
		return TextBlock(block)
	}
}
