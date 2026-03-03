//go:build !js || !wasm

package store

import (
	"log"

	"github.com/Nu11ified/golem/state"
)

// LocalStorageMiddleware is a no-op in non-WASM builds.
// In WASM builds, it persists state to localStorage after each dispatch.
func LocalStorageMiddleware(s *state.Store, action state.Action, next func(state.Action)) {
	next(action)
}

// ServerSyncMiddleware is a no-op in non-WASM builds.
// In WASM builds, it debounces server sync after state changes.
func ServerSyncMiddleware(s *state.Store, action state.Action, next func(state.Action)) {
	next(action)
}

// LoggerMiddleware logs actions for debugging.
func LoggerMiddleware(s *state.Store, action state.Action, next func(state.Action)) {
	log.Printf("[Store] Action: %s", action.Type)
	next(action)
	log.Printf("[Store] Action complete: %s", action.Type)
}
