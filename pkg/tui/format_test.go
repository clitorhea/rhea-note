package tui

import (
	"testing"
)

func TestAutoFormat(t *testing.T) {
	content := "```json\n{\"a\":1}\n```"
	formatted := autoFormatNote(content)
	if formatted == content {
		t.Errorf("Failed to format: %q", formatted)
	} else {
		t.Logf("Formatted: %q", formatted)
	}
}
