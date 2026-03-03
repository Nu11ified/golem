//go:build !js || !wasm

package blocks

import (
	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/src/app/models"
)

// TextBlock renders a paragraph text block.
func TextBlock(block *models.Block) *dom.Element {
	return dom.Div(
		dom.Class("block block-text"),
		dom.Div(
			dom.Class("block-content"),
			dom.Text(block.Content),
		),
	)
}
