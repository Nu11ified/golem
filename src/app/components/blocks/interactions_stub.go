//go:build !js || !wasm

package blocks

import (
	"strings"
	"unicode"
)

// DetectBlockType checks if text starts with a markdown-style shortcut
// and returns the new block type and remaining content.
// Returns ("", "") if no shortcut is detected.
func DetectBlockType(text string) (newType, remainingContent string) {
	if text == "" {
		return "", ""
	}

	// Check for code block
	if text == "```" || strings.HasPrefix(text, "```") {
		return "code", ""
	}

	// Check for divider
	if text == "---" {
		return "divider", ""
	}

	// Check for headings (# , ## , ### )
	if strings.HasPrefix(text, "### ") {
		return "h3", strings.TrimPrefix(text, "### ")
	}
	if strings.HasPrefix(text, "## ") {
		return "h2", strings.TrimPrefix(text, "## ")
	}
	if strings.HasPrefix(text, "# ") {
		return "h1", strings.TrimPrefix(text, "# ")
	}

	// Check for bullet list (- or *)
	if strings.HasPrefix(text, "- ") {
		return "bullet", strings.TrimPrefix(text, "- ")
	}
	if strings.HasPrefix(text, "* ") {
		return "bullet", strings.TrimPrefix(text, "* ")
	}

	// Check for numbered list (digit(s) followed by ". ")
	if len(text) >= 3 && unicode.IsDigit(rune(text[0])) {
		for i, ch := range text {
			if ch == '.' && i > 0 && i+1 < len(text) && text[i+1] == ' ' {
				return "numbered", text[i+2:]
			}
			if !unicode.IsDigit(ch) {
				break
			}
		}
	}

	return "", ""
}
