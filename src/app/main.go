//go:build js && wasm

package main

import (
	"github.com/Nu11ified/golem/dom"
)

func main() {
	app := dom.Div(
		dom.Class("app"),
		dom.H1("Golem Framework"),
		dom.P("App loaded successfully."),
	)
	dom.Render(app, "#app")

	// Keep the Go runtime alive
	select {}
}
