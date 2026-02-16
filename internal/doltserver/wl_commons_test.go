package doltserver

import (
	"os"
	"strings"
	"testing"
)

func TestGenerateWantedID(t *testing.T) {
	id := GenerateWantedID("Fix auth bug")
	if !strings.HasPrefix(id, "w-") {
		t.Errorf("GenerateWantedID() = %q, want prefix 'w-'", id)
	}
	// w- prefix + 10 hex chars = 12 chars total
	if len(id) != 12 {
		t.Errorf("GenerateWantedID() = %q, want 12 chars (got %d)", id, len(id))
	}
}

func TestGenerateWantedID_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := GenerateWantedID("same title")
		if seen[id] {
			t.Fatalf("GenerateWantedID produced duplicate: %s", id)
		}
		seen[id] = true
	}
}

func TestGetTownHandle_DoltHubOrg(t *testing.T) {
	orig := os.Getenv("DOLTHUB_ORG")
	defer os.Setenv("DOLTHUB_ORG", orig)

	os.Setenv("DOLTHUB_ORG", "my-org")
	got := GetTownHandle("/some/path/mytown")
	if got != "my-org" {
		t.Errorf("GetTownHandle() = %q, want 'my-org'", got)
	}
}

func TestGetTownHandle_Fallback(t *testing.T) {
	orig := os.Getenv("DOLTHUB_ORG")
	defer os.Setenv("DOLTHUB_ORG", orig)

	os.Unsetenv("DOLTHUB_ORG")
	got := GetTownHandle("/some/path/mytown")
	if got != "mytown" {
		t.Errorf("GetTownHandle() = %q, want 'mytown'", got)
	}
}

func TestWantedItem_Validation(t *testing.T) {
	// InsertWanted requires ID and Title - test that empty values error.
	// We can't run actual SQL here (no Dolt server), but we can test the
	// validation logic at the start of InsertWanted.
	err := InsertWanted("/nonexistent", &WantedItem{})
	if err == nil {
		t.Error("InsertWanted with empty ID should error")
	}
	if !strings.Contains(err.Error(), "ID cannot be empty") {
		t.Errorf("InsertWanted error = %q, want 'ID cannot be empty'", err)
	}

	err = InsertWanted("/nonexistent", &WantedItem{ID: "w-abc"})
	if err == nil {
		t.Error("InsertWanted with empty title should error")
	}
	if !strings.Contains(err.Error(), "title cannot be empty") {
		t.Errorf("InsertWanted error = %q, want 'title cannot be empty'", err)
	}
}
