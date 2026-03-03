//go:build js && wasm

package blocks

import (
	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/src/app/models"
)

// ToggleBlock renders a collapsible toggle block.
func ToggleBlock(block *models.Block) *dom.Element {
	return dom.Div(
		dom.Class("block block-toggle"),
		dom.Div(dom.Class("toggle-indicator"), dom.Text("\u25b6")),
		dom.Div(dom.Class("block-content"), dom.Text(block.Content)),
	)
}
