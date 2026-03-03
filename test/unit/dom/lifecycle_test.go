//go:build !js || !wasm

package dom_test

import (
	"testing"

	"github.com/Nu11ified/golem/dom"
)

func TestOnMount_StoredOnElement(t *testing.T) {
	called := false
	fn := func() func() {
		called = true
		return nil
	}

	el := dom.Div(dom.OnMount(fn))

	if len(el.MountCallbacks) != 1 {
		t.Fatalf("expected 1 mount callback, got %d", len(el.MountCallbacks))
	}

	// Verify the callback is the one we registered by invoking it
	el.MountCallbacks[0]()
	if !called {
		t.Error("expected mount callback to be the registered function")
	}
}

func TestOnUnmount_StoredOnElement(t *testing.T) {
	called := false
	fn := func() {
		called = true
	}

	el := dom.Div(dom.OnUnmount(fn))

	if len(el.UnmountCallbacks) != 1 {
		t.Fatalf("expected 1 unmount callback, got %d", len(el.UnmountCallbacks))
	}

	// Verify the callback is the one we registered by invoking it
	el.UnmountCallbacks[0]()
	if !called {
		t.Error("expected unmount callback to be the registered function")
	}
}

func TestOnUpdate_StoredOnElement(t *testing.T) {
	called := false
	fn := func() {
		called = true
	}

	el := dom.Div(dom.OnUpdate(fn))

	if len(el.UpdateCallbacks) != 1 {
		t.Fatalf("expected 1 update callback, got %d", len(el.UpdateCallbacks))
	}

	// Verify the callback is the one we registered by invoking it
	el.UpdateCallbacks[0]()
	if !called {
		t.Error("expected update callback to be the registered function")
	}
}

func TestMultipleLifecycleHooks(t *testing.T) {
	mountCount := 0
	unmountCount := 0

	mount1 := func() func() { mountCount++; return nil }
	mount2 := func() func() { mountCount++; return nil }
	unmount1 := func() { unmountCount++ }
	unmount2 := func() { unmountCount++ }

	el := dom.Div(
		dom.OnMount(mount1),
		dom.OnMount(mount2),
		dom.OnUnmount(unmount1),
		dom.OnUnmount(unmount2),
	)

	if len(el.MountCallbacks) != 2 {
		t.Fatalf("expected 2 mount callbacks, got %d", len(el.MountCallbacks))
	}
	if len(el.UnmountCallbacks) != 2 {
		t.Fatalf("expected 2 unmount callbacks, got %d", len(el.UnmountCallbacks))
	}

	// Invoke all and verify
	for _, cb := range el.MountCallbacks {
		cb()
	}
	if mountCount != 2 {
		t.Errorf("expected mountCount 2, got %d", mountCount)
	}

	for _, cb := range el.UnmountCallbacks {
		cb()
	}
	if unmountCount != 2 {
		t.Errorf("expected unmountCount 2, got %d", unmountCount)
	}
}

func TestOnMount_CleanupFunction(t *testing.T) {
	cleanupCalled := false
	fn := func() func() {
		return func() {
			cleanupCalled = true
		}
	}

	el := dom.Div(dom.OnMount(fn))

	if len(el.MountCallbacks) != 1 {
		t.Fatalf("expected 1 mount callback, got %d", len(el.MountCallbacks))
	}

	// Invoke mount callback and capture cleanup
	cleanup := el.MountCallbacks[0]()
	if cleanup == nil {
		t.Fatal("expected mount callback to return a cleanup function")
	}

	// Invoke cleanup
	cleanup()
	if !cleanupCalled {
		t.Error("expected cleanup function to have been called")
	}
}

func TestLifecycleHooks_WithOtherAttributes(t *testing.T) {
	mountCalled := false
	unmountCalled := false
	updateCalled := false

	el := dom.Div(
		dom.Class("container"),
		dom.Id("main"),
		dom.Text("hello"),
		dom.OnMount(func() func() {
			mountCalled = true
			return nil
		}),
		dom.OnUnmount(func() {
			unmountCalled = true
		}),
		dom.OnUpdate(func() {
			updateCalled = true
		}),
	)

	// Check that regular attributes are still set
	if el.Props["class"] != "container" {
		t.Errorf("expected class 'container', got %v", el.Props["class"])
	}
	if el.Props["id"] != "main" {
		t.Errorf("expected id 'main', got %v", el.Props["id"])
	}
	if el.Props["textContent"] != "hello" {
		t.Errorf("expected textContent 'hello', got %v", el.Props["textContent"])
	}

	// Check lifecycle hooks are stored
	if len(el.MountCallbacks) != 1 {
		t.Fatalf("expected 1 mount callback, got %d", len(el.MountCallbacks))
	}
	if len(el.UnmountCallbacks) != 1 {
		t.Fatalf("expected 1 unmount callback, got %d", len(el.UnmountCallbacks))
	}
	if len(el.UpdateCallbacks) != 1 {
		t.Fatalf("expected 1 update callback, got %d", len(el.UpdateCallbacks))
	}

	// Verify callbacks work
	el.MountCallbacks[0]()
	el.UnmountCallbacks[0]()
	el.UpdateCallbacks[0]()

	if !mountCalled {
		t.Error("mount callback was not invoked")
	}
	if !unmountCalled {
		t.Error("unmount callback was not invoked")
	}
	if !updateCalled {
		t.Error("update callback was not invoked")
	}
}
