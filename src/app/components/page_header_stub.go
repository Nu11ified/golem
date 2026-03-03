//go:build !js || !wasm

package components

import (
	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/src/app/models"
)

// CommonEmojis is the set of emojis available in the picker.
var CommonEmojis = []string{
	"📄", "📝", "📋", "📌", "📎", "📐", "📊", "📈",
	"🏠", "🔧", "💡", "⭐", "❤️", "🎯", "🚀", "🎨",
	"📦", "🔒", "🔑", "🎵", "📷", "🌍", "☀️", "🌙",
}

// PageHeader renders the page icon and editable title.
func PageHeader(page *models.Page) *dom.Element {
	title := page.Title
	if title == "" {
		title = "Untitled"
	}

	icon := dom.Div(
		dom.Class("page-header-icon"),
		dom.Text(page.Icon),
	)

	titleEl := dom.Div(
		dom.Class("page-header-title"),
		dom.Text(title),
	)

	return dom.Div(
		dom.Class("page-header"),
		icon,
		titleEl,
	)
}

// EmojiPicker renders a grid of emoji options.
func EmojiPicker() *dom.Element {
	var children []interface{}
	children = append(children, dom.Class("emoji-picker"))

	for _, emoji := range CommonEmojis {
		item := dom.Div(
			dom.Class("emoji-picker-item"),
			dom.Text(emoji),
		)
		children = append(children, item)
	}

	return dom.Div(children...)
}
