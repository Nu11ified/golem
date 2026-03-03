//go:build js && wasm

package store

import (
	"log"
	"sync"
	"time"

	"github.com/Nu11ified/golem/state"
)

var (
	syncTimer *time.Timer
	syncMu    sync.Mutex
)

// LocalStorageMiddleware persists workspace and editor state to localStorage after each dispatch.
func LocalStorageMiddleware(s *state.Store, action state.Action, next func(state.Action)) {
	next(action)

	// After reducers run, persist to localStorage
	p := state.NewPersistence()
	ws := s.GetState("workspace")
	if ws != nil {
		p.SaveState("notion_workspace", ws)
	}
	ed := s.GetState("editor")
	if ed != nil {
		p.SaveState("notion_editor", ed)
	}
}

// ServerSyncMiddleware debounces server sync 500ms after last dispatch.
func ServerSyncMiddleware(s *state.Store, action state.Action, next func(state.Action)) {
	next(action)

	// Debounce server sync
	syncMu.Lock()
	defer syncMu.Unlock()

	if syncTimer != nil {
		syncTimer.Stop()
	}

	syncTimer = time.AfterFunc(500*time.Millisecond, func() {
		syncWorkspaceToServer(s)
	})
}

func syncWorkspaceToServer(s *state.Store) {
	// Server sync will be wired up when grpc client is available
	// For now, this is a placeholder
	log.Println("[ServerSync] Would sync to server")
}

// LoggerMiddleware logs actions for debugging.
func LoggerMiddleware(s *state.Store, action state.Action, next func(state.Action)) {
	log.Printf("[Store] Action: %s", action.Type)
	next(action)
	log.Printf("[Store] Action complete: %s", action.Type)
}
