//go:build !js || !wasm

package blocks

import (
	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/src/app/models"
)

// BulletListBlock renders a bulleted list item block.
func BulletListBlock(block *models.Block) *dom.Element {
	return dom.Div(
		dom.Class("block block-bullet"),
		dom.Div(dom.Class("block-bullet-marker"), dom.Text("\u2022")),
		dom.Div(dom.Class("block-content"), dom.Text(block.Content)),
	)
}

// NumberedListBlock renders a numbered list item block.
func NumberedListBlock(block *models.Block) *dom.Element {
	return dom.Div(
		dom.Class("block block-numbered"),
		dom.Div(dom.Class("block-number"), dom.Text("1")),
		dom.Div(dom.Class("block-content"), dom.Text(block.Content)),
	)
}
