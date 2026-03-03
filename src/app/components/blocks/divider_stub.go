//go:build !js || !wasm

package blocks

import (
	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/src/app/models"
)

// DividerBlock renders a horizontal rule divider.
func DividerBlock(block *models.Block) *dom.Element {
	return dom.Div(
		dom.Class("block block-divider"),
		dom.El("hr"),
	)
}
