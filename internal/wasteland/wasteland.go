// Package wasteland provides shared logic for the Wasteland federation protocol.
// It handles interactions with wl-commons (the shared wanted board on DoltHub).
package wasteland

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	// UpstreamCommons is the canonical upstream commons database on DoltHub.
	UpstreamCommons = "hop/wl-commons"

	// WantedIDPrefix is the prefix for wanted item IDs.
	WantedIDPrefix = "w-"
)

// GenerateWantedID generates a unique wanted item ID: w-<10-char-hash>.
// The hash is derived from title + timestamp + random bytes for uniqueness.
func GenerateWantedID(title string) string {
	randomBytes := make([]byte, 8)
	_, _ = rand.Read(randomBytes) // crypto/rand only fails on broken systems

	input := fmt.Sprintf("%s:%d:%x", title, time.Now().UnixNano(), randomBytes)
	hash := sha256.Sum256([]byte(input))
	hashStr := hex.EncodeToString(hash[:])[:10]

	return WantedIDPrefix + hashStr
}

// CloneCommons clones the upstream wl-commons database to a temporary directory.
// Returns the path to the cloned database directory.
// The caller is responsible for cleaning up (os.RemoveAll) when done.
func CloneCommons(commonsRepo string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "wl-commons-*")
	if err != nil {
		return "", fmt.Errorf("creating temp dir: %w", err)
	}

	// Clone into a subdirectory so dolt has a proper working dir
	cloneDir := filepath.Join(tmpDir, "wl-commons")

	cmd := exec.Command("dolt", "clone", commonsRepo, cloneDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", fmt.Errorf("dolt clone %s: %w (%s)", commonsRepo, err, strings.TrimSpace(string(output)))
	}

	return cloneDir, nil
}

// DoltSQL runs a SQL query against a Dolt database directory using `dolt sql -q`.
func DoltSQL(dbDir, query string) (string, error) {
	cmd := exec.Command("dolt", "sql", "-q", query)
	cmd.Dir = dbDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("dolt sql: %w (%s)", err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

// DoltCommit stages all changes and commits with the given message.
func DoltCommit(dbDir, message string) error {
	addCmd := exec.Command("dolt", "add", ".")
	addCmd.Dir = dbDir
	if output, err := addCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("dolt add: %w (%s)", err, strings.TrimSpace(string(output)))
	}

	commitCmd := exec.Command("dolt", "commit", "-m", message)
	commitCmd.Dir = dbDir
	output, err := commitCmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(output))
		lower := strings.ToLower(msg)
		if strings.Contains(lower, "nothing to commit") {
			return nil
		}
		return fmt.Errorf("dolt commit: %w (%s)", err, msg)
	}
	return nil
}

// DoltPush pushes the current branch to origin.
func DoltPush(dbDir string) error {
	cmd := exec.Command("dolt", "push", "origin", "main")
	cmd.Dir = dbDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("dolt push: %w (%s)", err, strings.TrimSpace(string(output)))
	}
	return nil
}

// GetTownHandle returns the current town's handle for use in posted_by fields.
// Reads from mayor/town.json via the workspace package.
func GetTownHandle(townRoot string) (string, error) {
	// Use the DoltHub org if available (that's the public handle)
	if org := os.Getenv("DOLTHUB_ORG"); org != "" {
		return org, nil
	}

	// Read town.json directly to avoid circular imports
	townConfigPath := filepath.Join(townRoot, "mayor", "town.json")
	if _, err := os.Stat(townConfigPath); err != nil {
		return "", fmt.Errorf("town config not found: %w", err)
	}

	// Fall back to town name from config
	// Read the file and extract name field
	data, err := os.ReadFile(townConfigPath) //nolint:gosec // G304: trusted config path
	if err != nil {
		return "", fmt.Errorf("reading town config: %w", err)
	}

	// Simple JSON extraction to avoid importing config package
	// Look for "name": "value"
	s := string(data)
	idx := strings.Index(s, `"name"`)
	if idx < 0 {
		return "", fmt.Errorf("no 'name' field in town config")
	}
	rest := s[idx+len(`"name"`):]
	// Skip `: "`
	start := strings.Index(rest, `"`)
	if start < 0 {
		return "", fmt.Errorf("malformed town config")
	}
	rest = rest[start+1:]
	end := strings.Index(rest, `"`)
	if end < 0 {
		return "", fmt.Errorf("malformed town config")
	}
	return rest[:end], nil
}

// EscapeSQL escapes a string for use in SQL single-quoted literals.
func EscapeSQL(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
