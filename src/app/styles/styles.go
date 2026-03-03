//go:build js && wasm

package styles

import "github.com/Nu11ified/golem/css"

// GlobalCSS returns the complete CSS string.
func GlobalCSS() string {
	return globalCSS
}

// InjectGlobalCSS injects the global styles into the document.
func InjectGlobalCSS() {
	css.InjectStyles(globalCSS)
}
