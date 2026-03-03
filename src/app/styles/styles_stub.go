//go:build !js || !wasm

package styles

// GlobalCSS returns the complete CSS string for the Notion clone.
// Used by SSR to embed in the HTML document and by tests to verify CSS content.
func GlobalCSS() string {
	return globalCSS
}
