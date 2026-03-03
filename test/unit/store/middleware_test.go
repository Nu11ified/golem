//go:build !js || !wasm

package store_test

import (
	"testing"

	"github.com/Nu11ified/golem/src/app/store"
	golemstate "github.com/Nu11ified/golem/state"
)

func TestLocalStorageMiddlewareCallsNext(t *testing.T) {
	called := false
	next := func(action golemstate.Action) {
		called = true
	}

	store.LocalStorageMiddleware(nil, golemstate.Action{Type: "TEST"}, next)

	if !called {
		t.Error("middleware should call next")
	}
}

func TestServerSyncMiddlewareCallsNext(t *testing.T) {
	called := false
	next := func(action golemstate.Action) {
		called = true
	}

	store.ServerSyncMiddleware(nil, golemstate.Action{Type: "TEST"}, next)

	if !called {
		t.Error("middleware should call next")
	}
}

func TestLoggerMiddlewareCallsNext(t *testing.T) {
	called := false
	next := func(action golemstate.Action) {
		called = true
	}

	store.LoggerMiddleware(nil, golemstate.Action{Type: "TEST"}, next)

	if !called {
		t.Error("middleware should call next")
	}
}
