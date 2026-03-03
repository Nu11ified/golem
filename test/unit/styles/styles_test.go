//go:build !js || !wasm

package styles_test

import (
	"strings"
	"testing"

	"github.com/Nu11ified/golem/src/app/styles"
)

func TestCSSContainsSidebarStyles(t *testing.T) {
	css := styles.GlobalCSS()
	if !strings.Contains(css, ".sidebar") {
		t.Error("missing .sidebar class")
	}
	if !strings.Contains(css, "--bg-sidebar") {
		t.Error("missing --bg-sidebar variable")
	}
}

func TestCSSContainsEditorStyles(t *testing.T) {
	css := styles.GlobalCSS()
	if !strings.Contains(css, ".editor-area") {
		t.Error("missing .editor-area class")
	}
	if !strings.Contains(css, ".block") {
		t.Error("missing .block class")
	}
}

func TestCSSContainsMobileStyles(t *testing.T) {
	css := styles.GlobalCSS()
	if !strings.Contains(css, "@media") {
		t.Error("missing media queries")
	}
	if !strings.Contains(css, "768px") {
		t.Error("missing 768px breakpoint")
	}
}

func TestCSSContainsRootVariables(t *testing.T) {
	css := styles.GlobalCSS()
	if !strings.Contains(css, ":root") {
		t.Error("missing :root selector")
	}
	if !strings.Contains(css, "--bg-primary") {
		t.Error("missing --bg-primary variable")
	}
	if !strings.Contains(css, "--text-primary") {
		t.Error("missing --text-primary variable")
	}
	if !strings.Contains(css, "--font-body") {
		t.Error("missing --font-body variable")
	}
}

func TestCSSContainsBlockTypeStyles(t *testing.T) {
	css := styles.GlobalCSS()
	blockClasses := []string{
		".block-text", ".block-h1", ".block-h2", ".block-h3",
		".block-bullet", ".block-numbered", ".block-toggle",
		".block-code", ".block-divider",
	}
	for _, cls := range blockClasses {
		if !strings.Contains(css, cls) {
			t.Errorf("missing block class: %s", cls)
		}
	}
}

func TestCSSContainsSlashMenuStyles(t *testing.T) {
	css := styles.GlobalCSS()
	if !strings.Contains(css, ".slash-menu") {
		t.Error("missing .slash-menu class")
	}
}

func TestCSSContainsPageTreeStyles(t *testing.T) {
	css := styles.GlobalCSS()
	if !strings.Contains(css, ".page-tree") {
		t.Error("missing .page-tree class")
	}
	if !strings.Contains(css, ".page-tree-item") {
		t.Error("missing .page-tree-item class")
	}
}

func TestCSSContainsTouchTargets(t *testing.T) {
	css := styles.GlobalCSS()
	if !strings.Contains(css, "44px") {
		t.Error("missing 44px touch target sizing")
	}
}
