//go:build !js || !wasm

package dev_test

import (
	"sync"
	"testing"
	"time"

	"github.com/Nu11ified/golem/internal/dev"
	"github.com/Nu11ified/golem/internal/hmr"
)

// mockBroadcaster records all HMR messages sent through it.
type mockBroadcaster struct {
	mu            sync.Mutex
	reloads       int
	moduleUpdates []string
	errors        []string
}

func (m *mockBroadcaster) SendReload() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reloads++
}

func (m *mockBroadcaster) SendModuleUpdate(module string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.moduleUpdates = append(m.moduleUpdates, module)
}

func (m *mockBroadcaster) SendError(msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors = append(m.errors, msg)
}

func (m *mockBroadcaster) getReloads() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.reloads
}

func (m *mockBroadcaster) getModuleUpdates() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]string, len(m.moduleUpdates))
	copy(result, m.moduleUpdates)
	return result
}

func (m *mockBroadcaster) getErrors() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]string, len(m.errors))
	copy(result, m.errors)
	return result
}

// mockCompiler records compilation requests and can be configured to fail.
type mockCompiler struct {
	mu           sync.Mutex
	compilations []string
	failWith     error
}

func (m *mockCompiler) CompileModule(modulePath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.compilations = append(m.compilations, modulePath)
	return m.failWith
}

func (m *mockCompiler) CompileAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.compilations = append(m.compilations, "__full__")
	return m.failWith
}

func (m *mockCompiler) getCompilations() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]string, len(m.compilations))
	copy(result, m.compilations)
	return result
}

func TestHMRManagerInitialization(t *testing.T) {
	events := make(chan dev.FileEvent, 10)
	splitter := hmr.NewModuleSplitter(".")
	broadcaster := &mockBroadcaster{}
	compiler := &mockCompiler{}

	mgr := dev.NewHMRManager(events, splitter, broadcaster, compiler)
	if mgr == nil {
		t.Fatal("NewHMRManager returned nil")
	}

	mgr.Start()
	defer mgr.Stop()

	// Verify the manager is running by ensuring it does not panic
	// and that no messages are sent when there are no events
	time.Sleep(50 * time.Millisecond)

	if broadcaster.getReloads() != 0 {
		t.Errorf("expected 0 reloads, got %d", broadcaster.getReloads())
	}
	if len(broadcaster.getModuleUpdates()) != 0 {
		t.Errorf("expected 0 module updates, got %d", len(broadcaster.getModuleUpdates()))
	}
}

func TestPageChangeTriggerModuleUpdate(t *testing.T) {
	events := make(chan dev.FileEvent, 10)
	splitter := hmr.NewModuleSplitter(".")
	broadcaster := &mockBroadcaster{}
	compiler := &mockCompiler{}

	mgr := dev.NewHMRManager(events, splitter, broadcaster, compiler)
	mgr.Start()
	defer mgr.Stop()

	// Send a page file change event
	events <- dev.FileEvent{
		Path: "src/pages/home/index.go",
		Op:   dev.OpModify,
	}

	// Wait for debounce + processing
	time.Sleep(250 * time.Millisecond)

	// Should trigger a module update, NOT a full reload
	if broadcaster.getReloads() != 0 {
		t.Errorf("expected 0 full reloads for page change, got %d", broadcaster.getReloads())
	}

	updates := broadcaster.getModuleUpdates()
	if len(updates) == 0 {
		t.Fatal("expected at least 1 module update for page change, got 0")
	}

	// The module path should reference the page directory
	if updates[0] != "src/pages/home" {
		t.Errorf("expected module update for 'src/pages/home', got %q", updates[0])
	}
}

func TestShellChangeTriggerFullReload(t *testing.T) {
	events := make(chan dev.FileEvent, 10)
	splitter := hmr.NewModuleSplitter(".")
	broadcaster := &mockBroadcaster{}
	compiler := &mockCompiler{}

	mgr := dev.NewHMRManager(events, splitter, broadcaster, compiler)
	mgr.Start()
	defer mgr.Stop()

	// Send a shell/app file change event
	events <- dev.FileEvent{
		Path: "src/app/main.go",
		Op:   dev.OpModify,
	}

	// Wait for debounce + processing
	time.Sleep(250 * time.Millisecond)

	// Should trigger a full reload, not a module update
	if broadcaster.getReloads() != 1 {
		t.Errorf("expected 1 full reload for shell change, got %d", broadcaster.getReloads())
	}

	if len(broadcaster.getModuleUpdates()) != 0 {
		t.Errorf("expected 0 module updates for shell change, got %d", len(broadcaster.getModuleUpdates()))
	}
}

func TestComponentChangeTriggerFullReload(t *testing.T) {
	events := make(chan dev.FileEvent, 10)
	splitter := hmr.NewModuleSplitter(".")
	broadcaster := &mockBroadcaster{}
	compiler := &mockCompiler{}

	mgr := dev.NewHMRManager(events, splitter, broadcaster, compiler)
	mgr.Start()
	defer mgr.Stop()

	// Send a component file change (components are shell-level)
	events <- dev.FileEvent{
		Path: "src/components/header.go",
		Op:   dev.OpModify,
	}

	// Wait for debounce + processing
	time.Sleep(250 * time.Millisecond)

	if broadcaster.getReloads() != 1 {
		t.Errorf("expected 1 full reload for component change, got %d", broadcaster.getReloads())
	}
}

func TestDebouncingRapidChanges(t *testing.T) {
	events := make(chan dev.FileEvent, 100)
	splitter := hmr.NewModuleSplitter(".")
	broadcaster := &mockBroadcaster{}
	compiler := &mockCompiler{}

	mgr := dev.NewHMRManager(events, splitter, broadcaster, compiler)
	mgr.Start()
	defer mgr.Stop()

	// Send multiple rapid changes to the same shell file
	for i := 0; i < 10; i++ {
		events <- dev.FileEvent{
			Path: "src/app/main.go",
			Op:   dev.OpModify,
		}
	}

	// Wait for debounce + processing
	time.Sleep(350 * time.Millisecond)

	// Should only trigger ONE reload, not 10 due to debouncing
	reloads := broadcaster.getReloads()
	if reloads != 1 {
		t.Errorf("expected 1 reload after debouncing 10 rapid changes, got %d", reloads)
	}
}

func TestBuildErrorNotification(t *testing.T) {
	events := make(chan dev.FileEvent, 10)
	splitter := hmr.NewModuleSplitter(".")
	broadcaster := &mockBroadcaster{}
	compiler := &mockCompiler{
		failWith: &mockBuildError{msg: "compilation failed: syntax error in main.go:42"},
	}

	mgr := dev.NewHMRManager(events, splitter, broadcaster, compiler)
	mgr.Start()
	defer mgr.Stop()

	// Send a file change that will trigger a build
	events <- dev.FileEvent{
		Path: "src/app/main.go",
		Op:   dev.OpModify,
	}

	// Wait for debounce + processing
	time.Sleep(250 * time.Millisecond)

	// Should NOT send a reload when build fails
	if broadcaster.getReloads() != 0 {
		t.Errorf("expected 0 reloads on build error, got %d", broadcaster.getReloads())
	}

	// Should send an error message
	errors := broadcaster.getErrors()
	if len(errors) == 0 {
		t.Fatal("expected at least 1 error notification on build failure, got 0")
	}

	if errors[0] != "compilation failed: syntax error in main.go:42" {
		t.Errorf("unexpected error message: %q", errors[0])
	}
}

func TestMultiplePageChangesCompileIndependently(t *testing.T) {
	events := make(chan dev.FileEvent, 10)
	splitter := hmr.NewModuleSplitter(".")
	broadcaster := &mockBroadcaster{}
	compiler := &mockCompiler{}

	mgr := dev.NewHMRManager(events, splitter, broadcaster, compiler)
	mgr.Start()
	defer mgr.Stop()

	// Send a change to page "home"
	events <- dev.FileEvent{
		Path: "src/pages/home/index.go",
		Op:   dev.OpModify,
	}

	// Wait for debounce + processing
	time.Sleep(250 * time.Millisecond)

	// Send a change to page "about"
	events <- dev.FileEvent{
		Path: "src/pages/about/index.go",
		Op:   dev.OpModify,
	}

	// Wait for debounce + processing
	time.Sleep(250 * time.Millisecond)

	// Should have two module updates, not full reloads
	if broadcaster.getReloads() != 0 {
		t.Errorf("expected 0 full reloads, got %d", broadcaster.getReloads())
	}

	updates := broadcaster.getModuleUpdates()
	if len(updates) != 2 {
		t.Fatalf("expected 2 module updates, got %d: %v", len(updates), updates)
	}

	// Each page should have been compiled independently
	compilations := compiler.getCompilations()
	if len(compilations) != 2 {
		t.Fatalf("expected 2 compilations, got %d: %v", len(compilations), compilations)
	}

	// Verify both pages were compiled
	foundHome := false
	foundAbout := false
	for _, c := range compilations {
		if c == "src/pages/home" {
			foundHome = true
		}
		if c == "src/pages/about" {
			foundAbout = true
		}
	}
	if !foundHome {
		t.Error("expected compilation for 'src/pages/home'")
	}
	if !foundAbout {
		t.Error("expected compilation for 'src/pages/about'")
	}
}

func TestShellChangeTriggersFullCompile(t *testing.T) {
	events := make(chan dev.FileEvent, 10)
	splitter := hmr.NewModuleSplitter(".")
	broadcaster := &mockBroadcaster{}
	compiler := &mockCompiler{}

	mgr := dev.NewHMRManager(events, splitter, broadcaster, compiler)
	mgr.Start()
	defer mgr.Stop()

	events <- dev.FileEvent{
		Path: "src/app/main.go",
		Op:   dev.OpModify,
	}

	time.Sleep(250 * time.Millisecond)

	compilations := compiler.getCompilations()
	if len(compilations) != 1 {
		t.Fatalf("expected 1 compilation, got %d: %v", len(compilations), compilations)
	}
	if compilations[0] != "__full__" {
		t.Errorf("expected full compilation '__full__', got %q", compilations[0])
	}
}

func TestPageBuildErrorSendsError(t *testing.T) {
	events := make(chan dev.FileEvent, 10)
	splitter := hmr.NewModuleSplitter(".")
	broadcaster := &mockBroadcaster{}
	compiler := &mockCompiler{
		failWith: &mockBuildError{msg: "page build failed"},
	}

	mgr := dev.NewHMRManager(events, splitter, broadcaster, compiler)
	mgr.Start()
	defer mgr.Stop()

	// Send a page change that will fail to build
	events <- dev.FileEvent{
		Path: "src/pages/home/index.go",
		Op:   dev.OpModify,
	}

	time.Sleep(250 * time.Millisecond)

	// Should send error, not module update
	if len(broadcaster.getModuleUpdates()) != 0 {
		t.Errorf("expected 0 module updates on build error, got %d", len(broadcaster.getModuleUpdates()))
	}

	errors := broadcaster.getErrors()
	if len(errors) == 0 {
		t.Fatal("expected error notification for page build failure")
	}
	if errors[0] != "page build failed" {
		t.Errorf("unexpected error message: %q", errors[0])
	}
}

func TestStopPreventsProcessing(t *testing.T) {
	events := make(chan dev.FileEvent, 10)
	splitter := hmr.NewModuleSplitter(".")
	broadcaster := &mockBroadcaster{}
	compiler := &mockCompiler{}

	mgr := dev.NewHMRManager(events, splitter, broadcaster, compiler)
	mgr.Start()
	mgr.Stop()

	// Send events after stop
	events <- dev.FileEvent{
		Path: "src/app/main.go",
		Op:   dev.OpModify,
	}

	time.Sleep(250 * time.Millisecond)

	// Nothing should have been processed
	if broadcaster.getReloads() != 0 {
		t.Errorf("expected 0 reloads after stop, got %d", broadcaster.getReloads())
	}
}

// mockBuildError implements the error interface for testing.
type mockBuildError struct {
	msg string
}

func (e *mockBuildError) Error() string {
	return e.msg
}
