//go:build !js || !wasm

package components

import (
	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/src/app/components/blocks"
	"github.com/Nu11ified/golem/src/app/models"
)

// RenderEditorPage renders the complete editor page with header and block list.
func RenderEditorPage(page *models.Page, es *models.EditorState) *dom.Element {
	return dom.Div(
		dom.Class("editor-page"),
		PageHeader(page),
		BlockList(es),
	)
}

// BlockList renders all blocks in order.
func BlockList(es *models.EditorState) *dom.Element {
	var children []interface{}
	children = append(children, dom.Class("block-list"))

	if len(es.BlockOrder) == 0 {
		placeholder := dom.Div(
			dom.Class("block-list-empty"),
			dom.Text("Press Enter to start writing..."),
		)
		children = append(children, placeholder)
	} else {
		for _, blockID := range es.BlockOrder {
			if block, ok := es.Blocks[blockID]; ok {
				children = append(children, blocks.RenderBlock(block))
			}
		}
	}

	return dom.Div(children...)
}
