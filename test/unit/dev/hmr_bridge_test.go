//go:build !js || !wasm

package dev_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Nu11ified/golem/internal/dev"
)

func TestHMRBridgeJS_Embedded(t *testing.T) {
	if len(dev.HMRBridgeJS) == 0 {
		t.Fatal("HMRBridgeJS should be non-empty after embed")
	}
}

func TestHMRBridgeJS_ContainsWebSocket(t *testing.T) {
	if !strings.Contains(dev.HMRBridgeJS, "new WebSocket") {
		t.Fatal("HMRBridgeJS should contain WebSocket connection code")
	}
	if !strings.Contains(dev.HMRBridgeJS, "ws.onmessage") {
		t.Fatal("HMRBridgeJS should contain ws.onmessage handler")
	}
	if !strings.Contains(dev.HMRBridgeJS, "ws.onclose") {
		t.Fatal("HMRBridgeJS should contain ws.onclose handler")
	}
}

func TestHMRBridgeJS_ContainsModuleUpdate(t *testing.T) {
	if !strings.Contains(dev.HMRBridgeJS, "handleModuleUpdate") {
		t.Fatal("HMRBridgeJS should contain handleModuleUpdate function")
	}
	if !strings.Contains(dev.HMRBridgeJS, "module_update") {
		t.Fatal("HMRBridgeJS should contain module_update message type check")
	}
	if !strings.Contains(dev.HMRBridgeJS, "__golem_hmr_swap") {
		t.Fatal("HMRBridgeJS should contain __golem_hmr_swap hook")
	}
	if !strings.Contains(dev.HMRBridgeJS, "WebAssembly.instantiate") {
		t.Fatal("HMRBridgeJS should contain WebAssembly.instantiate call")
	}
}

func TestHMRBridgeJS_ContainsErrorOverlay(t *testing.T) {
	if !strings.Contains(dev.HMRBridgeJS, "showErrorOverlay") {
		t.Fatal("HMRBridgeJS should contain showErrorOverlay function")
	}
	if !strings.Contains(dev.HMRBridgeJS, "hideErrorOverlay") {
		t.Fatal("HMRBridgeJS should contain hideErrorOverlay function")
	}
	if !strings.Contains(dev.HMRBridgeJS, "golem-error-overlay") {
		t.Fatal("HMRBridgeJS should contain the error overlay element ID")
	}
	if !strings.Contains(dev.HMRBridgeJS, "Golem Build Error") {
		t.Fatal("HMRBridgeJS should contain the error overlay title")
	}
}

func TestBroadcaster_SendModuleUpdate(t *testing.T) {
	b := dev.NewBroadcaster()

	ch := make(chan string, 1)
	b.AddClient(ch)

	b.SendModuleUpdate("pages/home", "/modules/home.wasm")

	select {
	case msg := <-ch:
		var parsed map[string]string
		if err := json.Unmarshal([]byte(msg), &parsed); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}
		if parsed["type"] != "module_update" {
			t.Fatalf("expected type 'module_update', got %q", parsed["type"])
		}
		if parsed["module"] != "pages/home" {
			t.Fatalf("expected module 'pages/home', got %q", parsed["module"])
		}
		if parsed["url"] != "/modules/home.wasm" {
			t.Fatalf("expected url '/modules/home.wasm', got %q", parsed["url"])
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for module_update message")
	}
}
