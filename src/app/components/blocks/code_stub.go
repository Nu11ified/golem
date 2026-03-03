//go:build !js || !wasm

package blocks

import (
	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/src/app/models"
)

// CodeBlock renders a code block with syntax highlighting container.
func CodeBlock(block *models.Block) *dom.Element {
	return dom.Div(
		dom.Class("block block-code"),
		dom.El("pre",
			dom.El("code", dom.Text(block.Content)),
		),
	)
}
