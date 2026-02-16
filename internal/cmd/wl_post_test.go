package cmd

import (
	"testing"
)

func TestWlPostCmd_FlagRegistered(t *testing.T) {
	// Verify the --title flag is registered and marked required
	flag := wlPostCmd.Flags().Lookup("title")
	if flag == nil {
		t.Fatal("expected --title flag to be registered")
	}

	// Check other key flags exist
	for _, name := range []string{"project", "type", "priority", "effort", "tags", "dry-run"} {
		if f := wlPostCmd.Flags().Lookup(name); f == nil {
			t.Errorf("expected --%s flag to be registered", name)
		}
	}
}

func TestWlPostCmd_ValidatesType(t *testing.T) {
	// Save and restore global flags
	origTitle := wlPostTitle
	origType := wlPostType
	origDryRun := wlPostDryRun
	defer func() {
		wlPostTitle = origTitle
		wlPostType = origType
		wlPostDryRun = origDryRun
	}()

	wlPostTitle = "Test"
	wlPostType = "invalid_type"
	wlPostDryRun = true

	err := runWlPost(nil, nil)
	if err == nil {
		t.Fatal("expected error for invalid type, got nil")
	}
	if want := `invalid type "invalid_type"`; !containsStr(err.Error(), want) {
		t.Errorf("error %q should contain %q", err.Error(), want)
	}
}

func TestWlPostCmd_ValidatesEffort(t *testing.T) {
	origTitle := wlPostTitle
	origType := wlPostType
	origEffort := wlPostEffort
	origDryRun := wlPostDryRun
	defer func() {
		wlPostTitle = origTitle
		wlPostType = origType
		wlPostEffort = origEffort
		wlPostDryRun = origDryRun
	}()

	wlPostTitle = "Test"
	wlPostType = "feature"
	wlPostEffort = "bogus"
	wlPostDryRun = true

	err := runWlPost(nil, nil)
	if err == nil {
		t.Fatal("expected error for invalid effort, got nil")
	}
	if want := `invalid effort "bogus"`; !containsStr(err.Error(), want) {
		t.Errorf("error %q should contain %q", err.Error(), want)
	}
}

func TestWlPostCmd_ValidatesPriority(t *testing.T) {
	origTitle := wlPostTitle
	origType := wlPostType
	origEffort := wlPostEffort
	origPriority := wlPostPriority
	origDryRun := wlPostDryRun
	defer func() {
		wlPostTitle = origTitle
		wlPostType = origType
		wlPostEffort = origEffort
		wlPostPriority = origPriority
		wlPostDryRun = origDryRun
	}()

	wlPostTitle = "Test"
	wlPostType = "feature"
	wlPostEffort = "medium"
	wlPostPriority = 5
	wlPostDryRun = true

	err := runWlPost(nil, nil)
	if err == nil {
		t.Fatal("expected error for invalid priority, got nil")
	}
	if want := "invalid priority 5"; !containsStr(err.Error(), want) {
		t.Errorf("error %q should contain %q", err.Error(), want)
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
