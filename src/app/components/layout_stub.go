//go:build !js || !wasm

package components

import "github.com/Nu11ified/golem/dom"

// AppLayout creates the main application layout with sidebar and editor area
func AppLayout(sidebar, content *dom.Element) *dom.Element {
	var children []interface{}
	children = append(children, dom.Class("app-layout"))
	if sidebar != nil {
		children = append(children, sidebar)
	}
	editorArea := dom.Div(dom.Class("editor-area"))
	if content != nil {
		editorArea = dom.Div(dom.Class("editor-area"), content)
	}
	children = append(children, editorArea)
	return dom.Div(children...)
}
