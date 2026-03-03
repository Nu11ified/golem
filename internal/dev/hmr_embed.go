//go:build !js || !wasm

package dev

import _ "embed"

//go:embed hmr_bridge.js
var HMRBridgeJS string
