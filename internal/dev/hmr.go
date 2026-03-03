//go:build !js || !wasm

package dev

import (
	"log"
	"sync"
	"time"

	"github.com/Nu11ified/golem/internal/hmr"
)

// HMRBroadcaster defines the interface for broadcasting HMR messages.
type HMRBroadcaster interface {
	SendReload()
	SendModuleUpdate(moduleName, url string)
	SendError(msg string)
}

// HMRCompiler defines the interface for compiling modules.
type HMRCompiler interface {
	CompileModule(modulePath string) error
	CompileAll() error
}

// HMRSplitter defines the interface for module splitting analysis.
type HMRSplitter interface {
	Analyze() ([]hmr.ModuleInfo, error)
	DetermineAffectedModules(changedFiles []string, allModules []hmr.ModuleInfo) []hmr.ModuleInfo
}

// HMRManager wires together the file watcher, module splitter, and broadcaster
// to provide hot module replacement during development.
type HMRManager struct {
	splitter    HMRSplitter
	broadcaster HMRBroadcaster
	compiler    HMRCompiler
	done        chan struct{}
	wg          sync.WaitGroup
	debounce    time.Duration

	mu      sync.Mutex
	pending []string
	timer   *time.Timer
}

// NewHMRManager creates a new HMRManager.
func NewHMRManager(
	splitter HMRSplitter,
	broadcaster HMRBroadcaster,
	compiler HMRCompiler,
) *HMRManager {
	return &HMRManager{
		splitter:    splitter,
		broadcaster: broadcaster,
		compiler:    compiler,
		done:        make(chan struct{}),
		debounce:    100 * time.Millisecond,
	}
}

// HandleFileChange is called when a file change is detected.
func (m *HMRManager) HandleFileChange(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case <-m.done:
		return
	default:
	}

	m.pending = append(m.pending, path)

	if m.timer != nil {
		m.timer.Stop()
	}
	m.timer = time.AfterFunc(m.debounce, m.processPending)
}

// Stop stops the HMR manager.
func (m *HMRManager) Stop() {
	close(m.done)
	m.mu.Lock()
	if m.timer != nil {
		m.timer.Stop()
	}
	m.mu.Unlock()
	m.wg.Wait()
}

// processPending processes accumulated file changes.
func (m *HMRManager) processPending() {
	m.wg.Add(1)
	defer m.wg.Done()

	m.mu.Lock()
	files := m.pending
	m.pending = nil
	m.mu.Unlock()

	if len(files) == 0 {
		return
	}

	allModules, err := m.splitter.Analyze()
	if err != nil {
		log.Printf("[HMR] Failed to analyze modules: %v", err)
		m.broadcaster.SendError(err.Error())
		return
	}

	affected := m.splitter.DetermineAffectedModules(files, allModules)

	needsFullReload := false
	var pageModules []hmr.ModuleInfo
	for _, mod := range affected {
		if mod.IsShell {
			needsFullReload = true
			break
		}
		pageModules = append(pageModules, mod)
	}

	if needsFullReload {
		m.handleShellChange()
		return
	}

	for _, mod := range pageModules {
		m.handlePageChange(mod.Name)
	}
}

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

func (m *HMRManager) handlePageChange(modulePath string) {
	log.Printf("[HMR] Page change detected for %s, recompiling module...", modulePath)

	if err := m.compiler.CompileModule(modulePath); err != nil {
		log.Printf("[HMR] Module build error for %s: %v", modulePath, err)
		m.broadcaster.SendError(err.Error())
		return
	}

	log.Printf("[HMR] Module %s recompiled, sending update", modulePath)
	m.broadcaster.SendModuleUpdate(modulePath, modulePath+".wasm")
}
