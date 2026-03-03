//go:build js && wasm

package components

import (
	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/src/app/components/blocks"
	"github.com/Nu11ified/golem/src/app/models"
	"github.com/Nu11ified/golem/src/app/store"
	"github.com/Nu11ified/golem/state"
)

// RenderEditorPage renders the complete editor page with header and block list.
// In WASM builds it subscribes to the store so the block list re-renders when
// the editor state changes.
func RenderEditorPage(page *models.Page, es *models.EditorState) *dom.Element {
	return dom.Div(
		dom.Class("editor-page"),
		PageHeader(page),
		BlockList(es),
	)
}

// BlockList renders all blocks in order. When the list is empty it shows a
// placeholder that creates the first text block on click.
func BlockList(es *models.EditorState) *dom.Element {
	var children []interface{}
	children = append(children, dom.Class("block-list"))

	if len(es.BlockOrder) == 0 {
		placeholder := dom.Div(
			dom.Class("block-list-empty"),
			dom.Text("Press Enter to start writing..."),
			dom.OnClick(func() {
				if AppStore != nil {
					AppStore.Dispatch(state.Action{
						Type:    store.AddBlock,
						Payload: map[string]interface{}{"type": "text"},
					})
				}
			}),
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
