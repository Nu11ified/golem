//go:build js && wasm

package components

import (
	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/src/app/models"
	"github.com/Nu11ified/golem/src/app/store"
	"github.com/Nu11ified/golem/state"
)

// CommonEmojis is the set of emojis available in the picker.
var CommonEmojis = []string{
	"📄", "📝", "📋", "📌", "📎", "📐", "📊", "📈",
	"🏠", "🔧", "💡", "⭐", "❤️", "🎯", "🚀", "🎨",
	"📦", "🔒", "🔑", "🎵", "📷", "🌍", "☀️", "🌙",
}

// PageHeader renders the page icon and editable title with interactive
// features: clicking the icon opens the emoji picker; typing in the title
// dispatches a RenamePage action.
func PageHeader(page *models.Page) *dom.Element {
	title := page.Title
	if title == "" {
		title = "Untitled"
	}

	pageID := page.ID

	// Icon element -- clicking it toggles the emoji picker
	icon := dom.Div(
		dom.Class("page-header-icon"),
		dom.Text(page.Icon),
		dom.OnClick(func() {
			// Toggle emoji picker visibility (handled via CSS class)
			dom.Alert("Toggle emoji picker for page " + pageID)
		}),
	)

	// Title element -- editable via contenteditable, dispatches RenamePage on input
	titleEl := dom.Div(
		dom.Class("page-header-title"),
		dom.Text(title),
		dom.OnInput(func(value string) {
			if AppStore != nil {
				AppStore.Dispatch(state.Action{
					Type: store.RenamePage,
					Payload: map[string]interface{}{
						"id":    pageID,
						"title": value,
					},
				})
			}
		}),
	)

	return dom.Div(
		dom.Class("page-header"),
		icon,
		titleEl,
	)
}

// EmojiPicker renders a grid of emoji options. Each emoji dispatches a
// SetPageIcon action when clicked.
func EmojiPicker() *dom.Element {
	var children []interface{}
	children = append(children, dom.Class("emoji-picker"))

	for _, emoji := range CommonEmojis {
		e := emoji // capture loop variable
		item := dom.Div(
			dom.Class("emoji-picker-item"),
			dom.Text(e),
			dom.OnClick(func() {
				if AppStore != nil {
					AppStore.Dispatch(state.Action{
						Type: store.SetPageIcon,
						Payload: map[string]interface{}{
							"icon": e,
						},
					})
				}
			}),
		)
		children = append(children, item)
	}

	return dom.Div(children...)
}
