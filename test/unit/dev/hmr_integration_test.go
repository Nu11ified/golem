//go:build !js || !wasm

package dev_test

import (
	"fmt"
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

func (m *mockBroadcaster) SendModuleUpdate(moduleName, url string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.moduleUpdates = append(m.moduleUpdates, moduleName)
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

// mockSplitter returns preconfigured module analysis results.
type mockSplitter struct {
	modules  []hmr.ModuleInfo
	affected map[string][]hmr.ModuleInfo // changedFile -> affected modules
}

func newMockSplitter() *mockSplitter {
	return &mockSplitter{
		modules: []hmr.ModuleInfo{
			{Name: "shell", IsShell: true, SourceFiles: []string{"src/app/main.go", "src/components/header.go"}},
			{Name: "page_home", IsShell: false, SourceFiles: []string{"src/app/home/page.go"}, RoutePath: "/home"},
			{Name: "page_about", IsShell: false, SourceFiles: []string{"src/app/about/page.go"}, RoutePath: "/about"},
		},
		affected: make(map[string][]hmr.ModuleInfo),
	}
}

func (s *mockSplitter) Analyze() ([]hmr.ModuleInfo, error) {
	return s.modules, nil
}

func (s *mockSplitter) DetermineAffectedModules(changedFiles []string, allModules []hmr.ModuleInfo) []hmr.ModuleInfo {
	var result []hmr.ModuleInfo
	seen := make(map[string]bool)

	for _, f := range changedFiles {
		// Check if file belongs to any page module
		found := false
		for _, m := range allModules {
			if m.IsShell {
				continue
			}
			for _, src := range m.SourceFiles {
				if src == f {
					if !seen[m.Name] {
						seen[m.Name] = true
						result = append(result, m)
					}
					found = true
					break
				}
			}
		}
		if !found {
			// Not in a page module -> shell
			for _, m := range allModules {
				if m.IsShell && !seen[m.Name] {
					seen[m.Name] = true
					result = append(result, m)
				}
			}
		}
	}
	return result
}

func TestHMRManagerInitialization(t *testing.T) {
	splitter := newMockSplitter()
	broadcaster := &mockBroadcaster{}
	compiler := &mockCompiler{}

	mgr := dev.NewHMRManager(splitter, broadcaster, compiler)
	if mgr == nil {
		t.Fatal("NewHMRManager returned nil")
	}
	defer mgr.Stop()

	time.Sleep(50 * time.Millisecond)

	if broadcaster.getReloads() != 0 {
		t.Errorf("expected 0 reloads, got %d", broadcaster.getReloads())
	}
	if len(broadcaster.getModuleUpdates()) != 0 {
		t.Errorf("expected 0 module updates, got %d", len(broadcaster.getModuleUpdates()))
	}
}

func TestPageChangeTriggerModuleUpdate(t *testing.T) {
	splitter := newMockSplitter()
	broadcaster := &mockBroadcaster{}
	compiler := &mockCompiler{}

	mgr := dev.NewHMRManager(splitter, broadcaster, compiler)
	defer mgr.Stop()

	mgr.HandleFileChange("src/app/home/page.go")

	time.Sleep(250 * time.Millisecond)

	if broadcaster.getReloads() != 0 {
		t.Errorf("expected 0 full reloads for page change, got %d", broadcaster.getReloads())
	}

	updates := broadcaster.getModuleUpdates()
	if len(updates) == 0 {
		t.Fatal("expected at least 1 module update for page change, got 0")
	}

	if updates[0] != "page_home" {
		t.Errorf("expected module update for 'page_home', got %q", updates[0])
	}
}

func TestShellChangeTriggerFullReload(t *testing.T) {
	splitter := newMockSplitter()
	broadcaster := &mockBroadcaster{}
	compiler := &mockCompiler{}

	mgr := dev.NewHMRManager(splitter, broadcaster, compiler)
	defer mgr.Stop()

	mgr.HandleFileChange("src/app/main.go")

	time.Sleep(250 * time.Millisecond)

	if broadcaster.getReloads() != 1 {
		t.Errorf("expected 1 full reload for shell change, got %d", broadcaster.getReloads())
	}

	if len(broadcaster.getModuleUpdates()) != 0 {
		t.Errorf("expected 0 module updates for shell change, got %d", len(broadcaster.getModuleUpdates()))
	}
}

func TestComponentChangeTriggerFullReload(t *testing.T) {
	splitter := newMockSplitter()
	broadcaster := &mockBroadcaster{}
	compiler := &mockCompiler{}

	mgr := dev.NewHMRManager(splitter, broadcaster, compiler)
	defer mgr.Stop()

	mgr.HandleFileChange("src/components/header.go")

	time.Sleep(250 * time.Millisecond)

	if broadcaster.getReloads() != 1 {
		t.Errorf("expected 1 full reload for component change, got %d", broadcaster.getReloads())
	}
}

func TestDebouncingRapidChanges(t *testing.T) {
	splitter := newMockSplitter()
	broadcaster := &mockBroadcaster{}
	compiler := &mockCompiler{}

	mgr := dev.NewHMRManager(splitter, broadcaster, compiler)
	defer mgr.Stop()

	for i := 0; i < 10; i++ {
		mgr.HandleFileChange("src/app/main.go")
	}

	time.Sleep(350 * time.Millisecond)

	reloads := broadcaster.getReloads()
	if reloads != 1 {
		t.Errorf("expected 1 reload after debouncing 10 rapid changes, got %d", reloads)
	}
}

func TestBuildErrorNotification(t *testing.T) {
	splitter := newMockSplitter()
	broadcaster := &mockBroadcaster{}
	compiler := &mockCompiler{
		failWith: fmt.Errorf("compilation failed: syntax error in main.go:42"),
	}

	mgr := dev.NewHMRManager(splitter, broadcaster, compiler)
	defer mgr.Stop()

	mgr.HandleFileChange("src/app/main.go")

	time.Sleep(250 * time.Millisecond)

	if broadcaster.getReloads() != 0 {
		t.Errorf("expected 0 reloads on build error, got %d", broadcaster.getReloads())
	}

	errors := broadcaster.getErrors()
	if len(errors) == 0 {
		t.Fatal("expected at least 1 error notification on build failure, got 0")
	}

	if errors[0] != "compilation failed: syntax error in main.go:42" {
		t.Errorf("unexpected error message: %q", errors[0])
	}
}

func TestMultiplePageChangesCompileIndependently(t *testing.T) {
	splitter := newMockSplitter()
	broadcaster := &mockBroadcaster{}
	compiler := &mockCompiler{}

	mgr := dev.NewHMRManager(splitter, broadcaster, compiler)
	defer mgr.Stop()

	mgr.HandleFileChange("src/app/home/page.go")
	time.Sleep(250 * time.Millisecond)

	mgr.HandleFileChange("src/app/about/page.go")
	time.Sleep(250 * time.Millisecond)

	if broadcaster.getReloads() != 0 {
		t.Errorf("expected 0 full reloads, got %d", broadcaster.getReloads())
	}

	updates := broadcaster.getModuleUpdates()
	if len(updates) != 2 {
		t.Fatalf("expected 2 module updates, got %d: %v", len(updates), updates)
	}

	compilations := compiler.getCompilations()
	if len(compilations) != 2 {
		t.Fatalf("expected 2 compilations, got %d: %v", len(compilations), compilations)
	}
}

func TestShellChangeTriggersFullCompile(t *testing.T) {
	splitter := newMockSplitter()
	broadcaster := &mockBroadcaster{}
	compiler := &mockCompiler{}

	mgr := dev.NewHMRManager(splitter, broadcaster, compiler)
	defer mgr.Stop()

	mgr.HandleFileChange("src/app/main.go")
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
	splitter := newMockSplitter()
	broadcaster := &mockBroadcaster{}
	compiler := &mockCompiler{
		failWith: fmt.Errorf("page build failed"),
	}

	mgr := dev.NewHMRManager(splitter, broadcaster, compiler)
	defer mgr.Stop()

	mgr.HandleFileChange("src/app/home/page.go")
	time.Sleep(250 * time.Millisecond)

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
	splitter := newMockSplitter()
	broadcaster := &mockBroadcaster{}
	compiler := &mockCompiler{}

	mgr := dev.NewHMRManager(splitter, broadcaster, compiler)
	mgr.Stop()

	mgr.HandleFileChange("src/app/main.go")
	time.Sleep(250 * time.Millisecond)

	if broadcaster.getReloads() != 0 {
		t.Errorf("expected 0 reloads after stop, got %d", broadcaster.getReloads())
	}
}
