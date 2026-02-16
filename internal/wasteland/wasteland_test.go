package wasteland

import (
	"strings"
	"testing"
)

func TestGenerateWantedID(t *testing.T) {
	id := GenerateWantedID("Test title")
	if !strings.HasPrefix(id, WantedIDPrefix) {
		t.Errorf("expected prefix %q, got %q", WantedIDPrefix, id)
	}
	// w- prefix + 10 hex chars = 12 total
	if len(id) != 12 {
		t.Errorf("expected length 12, got %d (%q)", len(id), id)
	}

	// Should be unique
	id2 := GenerateWantedID("Test title")
	if id == id2 {
		t.Errorf("expected unique IDs, got same: %q", id)
	}
}

func TestEscapeSQL(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"it's", "it''s"},
		{"a''b", "a''''b"},
		{"", ""},
	}
	for _, tt := range tests {
		got := EscapeSQL(tt.input)
		if got != tt.want {
			t.Errorf("EscapeSQL(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestGetTownHandle_EnvOverride(t *testing.T) {
	// When DOLTHUB_ORG is set, it should be used as the handle
	t.Setenv("DOLTHUB_ORG", "test-org")
	handle, err := GetTownHandle("/nonexistent/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if handle != "test-org" {
		t.Errorf("expected 'test-org', got %q", handle)
	}
}
