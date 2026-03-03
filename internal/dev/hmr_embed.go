package dev

import (
	_ "embed"
	"net/http"
)

//go:embed hmr_bridge.js
var hmrBridgeJS []byte

// ServeHMRBridge returns an HTTP handler that serves the HMR bridge JavaScript.
// It is served at /golem-hmr.js.
func ServeHMRBridge() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Write(hmrBridgeJS)
	}
}
