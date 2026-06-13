package tui

import (
	"testing"
)

func TestAutoFormat(t *testing.T) {
	input := `{"key": "value"}`
	formatted, _ := autoFormatNote(input)
	if formatted == input {
		t.Errorf("Failed to format: %q", formatted)
	} else {
		t.Logf("Formatted: %q", formatted)
	}
}
