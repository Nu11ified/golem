//go:build !js || !wasm

package blocks

import (
	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/src/app/models"
)

// HeadingBlock renders an h1, h2, or h3 heading block.
func HeadingBlock(block *models.Block) *dom.Element {
	className := "block block-" + block.Type

	return dom.Div(
		dom.Class(className),
		dom.Div(
			dom.Class("block-content"),
			dom.Text(block.Content),
		),
	)
}
