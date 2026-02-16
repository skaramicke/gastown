package cmd

import (
	"testing"
)

func TestBuildBrowseQuery_DefaultFlags(t *testing.T) {
	// Save and restore flag state
	origStatus := wlBrowseStatus
	origProject := wlBrowseProject
	origType := wlBrowseType
	origPriority := wlBrowsePriority
	origLimit := wlBrowseLimit
	defer func() {
		wlBrowseStatus = origStatus
		wlBrowseProject = origProject
		wlBrowseType = origType
		wlBrowsePriority = origPriority
		wlBrowseLimit = origLimit
	}()

	wlBrowseStatus = "open"
	wlBrowseProject = ""
	wlBrowseType = ""
	wlBrowsePriority = -1
	wlBrowseLimit = 50

	query := buildBrowseQuery()

	if query != "SELECT id, title, project, type, priority, posted_by, status, effort_level FROM wanted WHERE status = 'open' ORDER BY priority ASC, created_at DESC LIMIT 50" {
		t.Errorf("unexpected default query: %s", query)
	}
}

func TestBuildBrowseQuery_AllFilters(t *testing.T) {
	origStatus := wlBrowseStatus
	origProject := wlBrowseProject
	origType := wlBrowseType
	origPriority := wlBrowsePriority
	origLimit := wlBrowseLimit
	defer func() {
		wlBrowseStatus = origStatus
		wlBrowseProject = origProject
		wlBrowseType = origType
		wlBrowsePriority = origPriority
		wlBrowseLimit = origLimit
	}()

	wlBrowseStatus = "claimed"
	wlBrowseProject = "gastown"
	wlBrowseType = "bug"
	wlBrowsePriority = 0
	wlBrowseLimit = 10

	query := buildBrowseQuery()

	expected := "SELECT id, title, project, type, priority, posted_by, status, effort_level FROM wanted WHERE status = 'claimed' AND project = 'gastown' AND type = 'bug' AND priority = 0 ORDER BY priority ASC, created_at DESC LIMIT 10"
	if query != expected {
		t.Errorf("unexpected query:\ngot:  %s\nwant: %s", query, expected)
	}
}

func TestBuildBrowseQuery_NoStatusFilter(t *testing.T) {
	origStatus := wlBrowseStatus
	origProject := wlBrowseProject
	origType := wlBrowseType
	origPriority := wlBrowsePriority
	origLimit := wlBrowseLimit
	defer func() {
		wlBrowseStatus = origStatus
		wlBrowseProject = origProject
		wlBrowseType = origType
		wlBrowsePriority = origPriority
		wlBrowseLimit = origLimit
	}()

	wlBrowseStatus = ""
	wlBrowseProject = ""
	wlBrowseType = ""
	wlBrowsePriority = -1
	wlBrowseLimit = 50

	query := buildBrowseQuery()

	expected := "SELECT id, title, project, type, priority, posted_by, status, effort_level FROM wanted ORDER BY priority ASC, created_at DESC LIMIT 50"
	if query != expected {
		t.Errorf("unexpected query:\ngot:  %s\nwant: %s", query, expected)
	}
}

func TestEscapeSQLString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"it's", "it''s"},
		{"a'b'c", "a''b''c"},
		{"", ""},
	}
	for _, tt := range tests {
		got := escapeSQLString(tt.input)
		if got != tt.expected {
			t.Errorf("escapeSQLString(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFormatPriority(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"0", "P0"},
		{"1", "P1"},
		{"2", "P2"},
		{"3", "P3"},
		{"4", "P4"},
		{"5", "5"},
		{"", ""},
	}
	for _, tt := range tests {
		got := formatPriority(tt.input)
		if got != tt.expected {
			t.Errorf("formatPriority(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseCSV(t *testing.T) {
	input := `id,title,project
w-abc,Fix bug,gastown
w-def,"Title with, comma",beads`

	rows := parseCSV(input)
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}

	// Header
	if rows[0][0] != "id" || rows[0][1] != "title" || rows[0][2] != "project" {
		t.Errorf("unexpected header: %v", rows[0])
	}

	// Row 1
	if rows[1][0] != "w-abc" || rows[1][1] != "Fix bug" || rows[1][2] != "gastown" {
		t.Errorf("unexpected row 1: %v", rows[1])
	}

	// Row 2 (with quoted comma)
	if rows[2][0] != "w-def" || rows[2][1] != "Title with, comma" || rows[2][2] != "beads" {
		t.Errorf("unexpected row 2: %v", rows[2])
	}
}

func TestParseCSVLine_EscapedQuotes(t *testing.T) {
	line := `w-abc,"He said ""hello""",gastown`
	fields := parseCSVLine(line)

	if len(fields) != 3 {
		t.Fatalf("expected 3 fields, got %d: %v", len(fields), fields)
	}
	if fields[1] != `He said "hello"` {
		t.Errorf("unexpected field: %q", fields[1])
	}
}

func TestParseCSV_Empty(t *testing.T) {
	rows := parseCSV("")
	if len(rows) != 0 {
		t.Errorf("expected 0 rows for empty input, got %d", len(rows))
	}
}

func TestFindWLCommonsFork_NotFound(t *testing.T) {
	// Should return empty when no fork exists
	result := findWLCommonsFork("/nonexistent/path")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}
