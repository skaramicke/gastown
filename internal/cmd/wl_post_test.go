package cmd

import (
	"testing"
)

func TestWlPostCmd_RequiresTitle(t *testing.T) {
	// Verify --title is marked as required.
	flag := wlPostCmd.Flags().Lookup("title")
	if flag == nil {
		t.Fatal("wl post missing --title flag")
	}

	// Check the required annotation
	annotations := flag.Annotations
	if req, ok := annotations["cobra_annotation_bash_completion_one_required_flag"]; !ok || len(req) == 0 {
		t.Error("--title should be marked as required")
	}
}

func TestWlPostCmd_Flags(t *testing.T) {
	// Verify all flags are registered on the command.
	flags := wlPostCmd.Flags()

	expectedFlags := []string{"title", "description", "project", "type", "priority", "effort", "tags"}
	for _, name := range expectedFlags {
		if flags.Lookup(name) == nil {
			t.Errorf("wl post missing flag: --%s", name)
		}
	}
}

func TestWlPostCmd_DefaultValues(t *testing.T) {
	flags := wlPostCmd.Flags()

	// priority defaults to 2
	pFlag := flags.Lookup("priority")
	if pFlag.DefValue != "2" {
		t.Errorf("priority default = %q, want '2'", pFlag.DefValue)
	}

	// effort defaults to medium
	eFlag := flags.Lookup("effort")
	if eFlag.DefValue != "medium" {
		t.Errorf("effort default = %q, want 'medium'", eFlag.DefValue)
	}
}

func TestWlPostCmd_IsSubcommand(t *testing.T) {
	found := false
	for _, child := range wlCmd.Commands() {
		if child.Name() == "post" {
			found = true
			break
		}
	}
	if !found {
		t.Error("wl post should be a subcommand of wl")
	}
}

func TestWlCmd_GroupID(t *testing.T) {
	if wlCmd.GroupID != GroupWork {
		t.Errorf("wl GroupID = %q, want %q", wlCmd.GroupID, GroupWork)
	}
}

func TestWlCmd_IsRootSubcommand(t *testing.T) {
	found := false
	for _, child := range rootCmd.Commands() {
		if child.Name() == "wl" {
			found = true
			break
		}
	}
	if !found {
		t.Error("wl should be a subcommand of root")
	}
}
