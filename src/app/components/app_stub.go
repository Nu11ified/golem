//go:build !js || !wasm

package components

// InitApp is a no-op in non-WASM builds. SSR setup is done separately.
func InitApp() {}
