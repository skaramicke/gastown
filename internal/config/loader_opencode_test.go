//go:build !windows
// +build !windows

package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/steveyegge/gastown/internal/config"
)

func TestBuildAgentStartupCommand_OpenCodePlugin(t *testing.T) {
	townRoot := t.TempDir()

	// Create town structure
	if err := os.MkdirAll(filepath.Join(townRoot, "mayor"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(townRoot, "settings"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(townRoot, "mayor", "town.json"), []byte(`{"name":"test-town"}`), 0644); err != nil {
		t.Fatal(err)
	}

	config.ResetRegistryForTesting()

	// Set opencode as the default agent
	settingsPath := config.TownSettingsPath(townRoot)
	settings, err := config.LoadOrCreateTownSettings(settingsPath)
	if err != nil {
		t.Fatalf("Failed to load town settings: %v", err)
	}
	settings.DefaultAgent = "opencode"
	if err := config.SaveTownSettings(settingsPath, settings); err != nil {
		t.Fatalf("Failed to save town settings: %v", err)
	}

	// Build the command for the mayor role
	cmd := config.BuildAgentStartupCommand("mayor", "", townRoot, "", "")

	// Check if the command contains the opencode plugin flags
	if !strings.Contains(cmd, "opencode") {
		t.Errorf("Expected command to contain 'opencode', but it didn't. Got: %s", cmd)
	}
	expectedPluginArgs := "--plugin-dir .opencode/plugins --plugin gastown.js"
	if !strings.Contains(cmd, expectedPluginArgs) {
		t.Errorf("Expected command to contain plugin args '%s', but it didn't. Got: %s", expectedPluginArgs, cmd)
	}

	// Check that the plugin file was created
	pluginPath := filepath.Join(townRoot, ".opencode", "plugins", "gastown.js")
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		t.Errorf("Expected plugin file to be created at %s, but it wasn't", pluginPath)
	}
}
