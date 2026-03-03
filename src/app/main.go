//go:build js && wasm

package main

import (
	"github.com/Nu11ified/golem/src/app/components"
)

func main() {
	components.InitApp()

	// Keep the Go runtime alive
	select {}
}
