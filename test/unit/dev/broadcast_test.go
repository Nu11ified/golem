//go:build !js || !wasm

package dev_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Nu11ified/golem/internal/dev"
)

func TestBroadcaster_AddRemoveClient(t *testing.T) {
	b := dev.NewBroadcaster()

	ch := make(chan string, 1)
	id := b.AddClient(ch)

	if b.ClientCount() != 1 {
		t.Fatalf("expected 1 client, got %d", b.ClientCount())
	}

	b.RemoveClient(id)

	if b.ClientCount() != 0 {
		t.Fatalf("expected 0 clients after removal, got %d", b.ClientCount())
	}
}

func TestBroadcaster_SendReload(t *testing.T) {
	b := dev.NewBroadcaster()

	ch := make(chan string, 1)
	b.AddClient(ch)

	b.SendReload()

	select {
	case msg := <-ch:
		if msg != "reload" {
			t.Fatalf("expected 'reload', got %q", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for reload message")
	}
}

func TestBroadcaster_SendError(t *testing.T) {
	b := dev.NewBroadcaster()

	ch := make(chan string, 1)
	b.AddClient(ch)

	b.SendError("build failed")

	select {
	case msg := <-ch:
		var parsed map[string]string
		if err := json.Unmarshal([]byte(msg), &parsed); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}
		if parsed["type"] != "error" {
			t.Fatalf("expected type 'error', got %q", parsed["type"])
		}
		if parsed["message"] != "build failed" {
			t.Fatalf("expected message 'build failed', got %q", parsed["message"])
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for error message")
	}
}

func TestBroadcaster_MultipleClients(t *testing.T) {
	b := dev.NewBroadcaster()

	ch1 := make(chan string, 1)
	ch2 := make(chan string, 1)

	b.AddClient(ch1)
	b.AddClient(ch2)

	if b.ClientCount() != 2 {
		t.Fatalf("expected 2 clients, got %d", b.ClientCount())
	}

	b.SendReload()

	for i, ch := range []chan string{ch1, ch2} {
		select {
		case msg := <-ch:
			if msg != "reload" {
				t.Fatalf("client %d: expected 'reload', got %q", i+1, msg)
			}
		case <-time.After(time.Second):
			t.Fatalf("client %d: timeout waiting for reload message", i+1)
		}
	}
}
