//go:build !js || !wasm

package integration_test

import (
	"strings"
	"testing"

	"github.com/Nu11ified/golem/internal/config"
	"github.com/Nu11ified/golem/internal/dev"
)

func TestGenerateDevHTML(t *testing.T) {
	cfg := &config.Config{
		ProjectName: "TestProject",
		Dev: config.DevConfig{
			Port:      8080,
			HotReload: true,
		},
	}

	s := dev.NewServer(cfg)
	html := s.GenerateDevHTML()

	required := []string{
		`<div id="app">`,
		"wasm_exec.js",
		"app.wasm",
		"TestProject",
	}

	for _, want := range required {
		if !strings.Contains(html, want) {
			t.Errorf("generated HTML missing expected content %q", want)
		}
	}
}
