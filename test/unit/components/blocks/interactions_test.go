//go:build !js || !wasm

package blocks_test

import (
	"testing"

	"github.com/Nu11ified/golem/src/app/components/blocks"
)

func TestDetectBlockTypeHeading1(t *testing.T) {
	newType, content := blocks.DetectBlockType("# Hello")
	if newType != "h1" {
		t.Errorf("expected h1, got %q", newType)
	}
	if content != "Hello" {
		t.Errorf("expected 'Hello', got %q", content)
	}
}

func TestDetectBlockTypeHeading2(t *testing.T) {
	newType, content := blocks.DetectBlockType("## World")
	if newType != "h2" {
		t.Errorf("expected h2, got %q", newType)
	}
	if content != "World" {
		t.Errorf("expected 'World', got %q", content)
	}
}

func TestDetectBlockTypeHeading3(t *testing.T) {
	newType, content := blocks.DetectBlockType("### Sub heading")
	if newType != "h3" {
		t.Errorf("expected h3, got %q", newType)
	}
	if content != "Sub heading" {
		t.Errorf("expected 'Sub heading', got %q", content)
	}
}

func TestDetectBlockTypeBulletDash(t *testing.T) {
	newType, content := blocks.DetectBlockType("- Bullet item")
	if newType != "bullet" {
		t.Errorf("expected bullet, got %q", newType)
	}
	if content != "Bullet item" {
		t.Errorf("expected 'Bullet item', got %q", content)
	}
}

func TestDetectBlockTypeBulletAsterisk(t *testing.T) {
	newType, content := blocks.DetectBlockType("* Star bullet")
	if newType != "bullet" {
		t.Errorf("expected bullet, got %q", newType)
	}
	if content != "Star bullet" {
		t.Errorf("expected 'Star bullet', got %q", content)
	}
}

func TestDetectBlockTypeNumbered(t *testing.T) {
	newType, content := blocks.DetectBlockType("1. First item")
	if newType != "numbered" {
		t.Errorf("expected numbered, got %q", newType)
	}
	if content != "First item" {
		t.Errorf("expected 'First item', got %q", content)
	}
}

func TestDetectBlockTypeDivider(t *testing.T) {
	newType, content := blocks.DetectBlockType("---")
	if newType != "divider" {
		t.Errorf("expected divider, got %q", newType)
	}
	if content != "" {
		t.Errorf("expected empty content, got %q", content)
	}
}

func TestDetectBlockTypeCode(t *testing.T) {
	newType, content := blocks.DetectBlockType("```")
	if newType != "code" {
		t.Errorf("expected code, got %q", newType)
	}
	if content != "" {
		t.Errorf("expected empty content, got %q", content)
	}
}

func TestDetectBlockTypeNoMatch(t *testing.T) {
	newType, content := blocks.DetectBlockType("Hello world")
	if newType != "" {
		t.Errorf("expected empty type, got %q", newType)
	}
	if content != "" {
		t.Errorf("expected empty content, got %q", content)
	}
}

func TestDetectBlockTypeJustHash(t *testing.T) {
	// Just "#" without space should not convert
	newType, _ := blocks.DetectBlockType("#noSpace")
	if newType != "" {
		t.Errorf("expected no match for '#noSpace', got %q", newType)
	}
}

func TestDetectBlockTypeNumberedMultiDigit(t *testing.T) {
	newType, content := blocks.DetectBlockType("12. Twelfth item")
	if newType != "numbered" {
		t.Errorf("expected numbered, got %q", newType)
	}
	if content != "Twelfth item" {
		t.Errorf("expected 'Twelfth item', got %q", content)
	}
}

func TestDetectBlockTypeEmptyString(t *testing.T) {
	newType, content := blocks.DetectBlockType("")
	if newType != "" || content != "" {
		t.Errorf("empty string should return empty: got type=%q content=%q", newType, content)
	}
}
