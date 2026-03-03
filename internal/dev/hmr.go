package dev

import (
	"log"
	"sync"
	"time"

	"github.com/Nu11ified/golem/internal/hmr"
)

// HMRBroadcaster defines the interface for broadcasting HMR messages.
// This allows the HMRManager to be tested with mock implementations.
type HMRBroadcaster interface {
	SendReload()
	SendModuleUpdate(module string)
	SendError(msg string)
}

// HMRCompiler defines the interface for compiling modules.
// This allows the HMRManager to be tested with mock implementations.
type HMRCompiler interface {
	// CompileModule compiles a single page module identified by modulePath.
	CompileModule(modulePath string) error
	// CompileAll performs a full rebuild of all modules.
	CompileAll() error
}

// HMRManager wires together the file watcher, module splitter, and broadcaster
// to provide hot module replacement during development.
//
// When a file change is detected:
//  1. The event is debounced (100ms window) to batch rapid changes.
//  2. The ModuleSplitter determines if the change is a shell or page change.
//  3. For page changes: only that page module is recompiled, then a module
//     update message is sent via the broadcaster.
//  4. For shell changes: a full rebuild is performed, then a full reload
//     message is sent.
//  5. On build errors: an error message is sent via the broadcaster.
type HMRManager struct {
	events      <-chan FileEvent
	splitter    *hmr.ModuleSplitter
	broadcaster HMRBroadcaster
	compiler    HMRCompiler
	done        chan struct{}
	wg          sync.WaitGroup
	debounce    time.Duration
}

// NewHMRManager creates a new HMRManager that listens for file events on the
// provided channel and uses the splitter, broadcaster, and compiler to handle
// hot module replacement.
func NewHMRManager(
	events <-chan FileEvent,
	splitter *hmr.ModuleSplitter,
	broadcaster HMRBroadcaster,
	compiler HMRCompiler,
) *HMRManager {
	return &HMRManager{
		events:      events,
		splitter:    splitter,
		broadcaster: broadcaster,
		compiler:    compiler,
		done:        make(chan struct{}),
		debounce:    100 * time.Millisecond,
	}
}

// Start begins processing file change events in the background.
func (m *HMRManager) Start() {
	m.wg.Add(1)
	go m.run()
}

// Stop stops processing file change events and waits for the background
// goroutine to finish.
func (m *HMRManager) Stop() {
	close(m.done)
	m.wg.Wait()
}

func (m *HMRManager) run() {
	defer m.wg.Done()

	// pending accumulates events during the debounce window.
	var pending []FileEvent
	var debounceTimer *time.Timer
	var debounceCh <-chan time.Time

	for {
		select {
		case <-m.done:
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			return

		case event, ok := <-m.events:
			if !ok {
				return
			}
			pending = append(pending, event)

			// Reset the debounce timer on every new event
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.NewTimer(m.debounce)
			debounceCh = debounceTimer.C

		case <-debounceCh:
			// Debounce window elapsed: process accumulated events
			m.processBatch(pending)
			pending = nil
			debounceCh = nil
		}
	}
}

// processBatch handles a batch of debounced file change events. It deduplicates
// affected modules and determines whether each needs a module update or full reload.
func (m *HMRManager) processBatch(events []FileEvent) {
	if len(events) == 0 {
		return
	}

	// Analyze all events and determine affected modules
	needsFullReload := false
	pageModules := make(map[string]bool)

	for _, event := range events {
		result := m.splitter.DetermineAffectedModules(event.Path)
		if result.IsShell {
			needsFullReload = true
			break // Shell change means full reload, no need to check more
		}
		// Page change: track the module path
		pageModules[result.ModulePath] = true
	}

	if needsFullReload {
		m.handleShellChange()
		return
	}

	// Handle page changes independently
	for modulePath := range pageModules {
		m.handlePageChange(modulePath)
	}
}

// handleShellChange performs a full rebuild and sends a reload message.
func (m *HMRManager) handleShellChange() {
	log.Println("[HMR] Shell change detected, performing full rebuild...")

	if err := m.compiler.CompileAll(); err != nil {
		log.Printf("[HMR] Build error: %v", err)
		m.broadcaster.SendError(err.Error())
		return
	}

	log.Println("[HMR] Full rebuild complete, sending reload")
	m.broadcaster.SendReload()
}

// handlePageChange compiles a single page module and sends a module update.
func (m *HMRManager) handlePageChange(modulePath string) {
	log.Printf("[HMR] Page change detected for %s, recompiling module...", modulePath)

	if err := m.compiler.CompileModule(modulePath); err != nil {
		log.Printf("[HMR] Module build error for %s: %v", modulePath, err)
		m.broadcaster.SendError(err.Error())
		return
	}

	log.Printf("[HMR] Module %s recompiled, sending update", modulePath)
	m.broadcaster.SendModuleUpdate(modulePath)
}
