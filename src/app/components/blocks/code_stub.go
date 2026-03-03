//go:build !js || !wasm

package blocks

import (
	"fmt"

	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/src/app/models"
)

// CodeBlock renders a code block with syntax highlighting container.
// If the block has a "language" prop, a language label is rendered above the code.
func CodeBlock(block *models.Block) *dom.Element {
	language := ""
	if block.Props != nil {
		if lang, ok := block.Props["language"]; ok {
			language = fmt.Sprintf("%v", lang)
		}
	}

	if language != "" {
		return dom.Div(
			dom.Class("block block-code"),
			dom.Div(dom.Class("code-language"), dom.Text(language)),
			dom.El("pre",
				dom.El("code", dom.Text(block.Content)),
			),
		)
	}

	return dom.Div(
		dom.Class("block block-code"),
		dom.El("pre",
			dom.El("code", dom.Text(block.Content)),
		),
	)
}
