package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/steveyegge/gastown/internal/style"
	"github.com/steveyegge/gastown/internal/workspace"
)

// wl command flags
var (
	wlBrowseProject  string
	wlBrowseStatus   string
	wlBrowseType     string
	wlBrowsePriority int
	wlBrowseLimit    int
	wlBrowseJSON     bool

	wlSyncDryRun bool

	// Default upstream org/database for wl-commons
	wlCommonsOrg = "hop"
	wlCommonsDB  = "wl-commons"
)

var wlCmd = &cobra.Command{
	Use:     "wl",
	GroupID: GroupWork,
	Short:   "Wasteland federation commands",
	RunE:    requireSubcommand,
	Long: `Interact with the Wasteland federation commons.

The Wasteland is a federated work board built on DoltHub. Towns post wanted
items (open work), claim them, complete them, and earn reputation stamps.

Data lives in the hop/wl-commons DoltHub database. Read operations use the
clone-then-discard pattern: clone to a temp directory, query, delete.

COMMANDS:
  browse    Browse wanted items on the commons board
  sync      Pull upstream changes into your local fork`,
}

var wlBrowseCmd = &cobra.Command{
	Use:   "browse",
	Short: "Browse wanted items on the commons board",
	Args:  cobra.NoArgs,
	RunE:  runWLBrowse,
	Long: `Browse the Wasteland wanted board (hop/wl-commons).

Uses the clone-then-discard pattern: clones the commons database to a
temporary directory, queries it, then deletes the clone.

EXAMPLES:
  gt wl browse                          # All open wanted items
  gt wl browse --project gastown        # Filter by project
  gt wl browse --type bug               # Only bugs
  gt wl browse --status claimed         # Claimed items
  gt wl browse --priority 0             # Critical priority only
  gt wl browse --limit 5               # Show 5 items
  gt wl browse --json                   # JSON output`,
}

var wlSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Pull upstream changes into local wl-commons fork",
	Args:  cobra.NoArgs,
	RunE:  runWLSync,
	Long: `Sync your local wl-commons fork with the upstream hop/wl-commons.

If you have a local fork of wl-commons (for posting wanted items or
completions), this pulls the latest changes from upstream.

Requires a local Dolt clone of your fork with 'upstream' remote configured.

EXAMPLES:
  gt wl sync                # Pull upstream changes
  gt wl sync --dry-run      # Show what would change`,
}

func init() {
	// Browse flags
	wlBrowseCmd.Flags().StringVar(&wlBrowseProject, "project", "", "Filter by project (e.g., gastown, beads, hop)")
	wlBrowseCmd.Flags().StringVar(&wlBrowseStatus, "status", "open", "Filter by status (open, claimed, in_review, completed, withdrawn)")
	wlBrowseCmd.Flags().StringVar(&wlBrowseType, "type", "", "Filter by type (feature, bug, design, rfc, docs)")
	wlBrowseCmd.Flags().IntVar(&wlBrowsePriority, "priority", -1, "Filter by priority (0=critical, 2=medium, 4=backlog)")
	wlBrowseCmd.Flags().IntVar(&wlBrowseLimit, "limit", 50, "Maximum items to display")
	wlBrowseCmd.Flags().BoolVar(&wlBrowseJSON, "json", false, "Output as JSON")

	// Sync flags
	wlSyncCmd.Flags().BoolVar(&wlSyncDryRun, "dry-run", false, "Show what would change without pulling")

	wlCmd.AddCommand(wlBrowseCmd)
	wlCmd.AddCommand(wlSyncCmd)
	rootCmd.AddCommand(wlCmd)
}

// runWLBrowse implements the clone-then-discard pattern for browsing wl-commons.
func runWLBrowse(cmd *cobra.Command, args []string) error {
	// Verify we're in a Gas Town workspace
	if _, err := workspace.FindFromCwdOrError(); err != nil {
		return fmt.Errorf("not in a Gas Town workspace: %w", err)
	}

	// Check dolt is available
	doltPath, err := exec.LookPath("dolt")
	if err != nil {
		return fmt.Errorf("dolt not found in PATH — install from https://docs.dolthub.com/introduction/installation")
	}

	// Create temp directory for clone
	tmpDir, err := os.MkdirTemp("", "wl-browse-*")
	if err != nil {
		return fmt.Errorf("creating temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	cloneDir := filepath.Join(tmpDir, wlCommonsDB)

	// Clone wl-commons from DoltHub
	remote := fmt.Sprintf("%s/%s", wlCommonsOrg, wlCommonsDB)
	fmt.Printf("Cloning %s...\n", style.Bold.Render(remote))

	cloneCmd := exec.Command(doltPath, "clone", remote, cloneDir)
	cloneCmd.Stderr = os.Stderr
	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("cloning %s: %w\nEnsure the database exists on DoltHub: https://www.dolthub.com/%s", remote, err, remote)
	}
	fmt.Printf("%s Cloned successfully\n\n", style.Bold.Render("✓"))

	// Build query
	query := buildBrowseQuery()

	if wlBrowseJSON {
		return queryJSON(doltPath, cloneDir, query)
	}

	return queryTable(doltPath, cloneDir, query)
}

// buildBrowseQuery constructs the SQL query based on flags.
func buildBrowseQuery() string {
	var conditions []string

	if wlBrowseStatus != "" {
		conditions = append(conditions, fmt.Sprintf("status = '%s'", escapeSQLString(wlBrowseStatus)))
	}
	if wlBrowseProject != "" {
		conditions = append(conditions, fmt.Sprintf("project = '%s'", escapeSQLString(wlBrowseProject)))
	}
	if wlBrowseType != "" {
		conditions = append(conditions, fmt.Sprintf("type = '%s'", escapeSQLString(wlBrowseType)))
	}
	if wlBrowsePriority >= 0 {
		conditions = append(conditions, fmt.Sprintf("priority = %d", wlBrowsePriority))
	}

	query := "SELECT id, title, project, type, priority, posted_by, status, effort_level FROM wanted"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY priority ASC, created_at DESC"
	query += fmt.Sprintf(" LIMIT %d", wlBrowseLimit)

	return query
}

// escapeSQLString escapes single quotes for safe SQL string interpolation.
func escapeSQLString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// queryJSON runs the query and outputs raw JSON.
func queryJSON(doltPath, cloneDir, query string) error {
	sqlCmd := exec.Command(doltPath, "sql", "-q", query, "-r", "json")
	sqlCmd.Dir = cloneDir
	sqlCmd.Stdout = os.Stdout
	sqlCmd.Stderr = os.Stderr
	return sqlCmd.Run()
}

// queryTable runs the query and renders a styled table.
func queryTable(doltPath, cloneDir, query string) error {
	sqlCmd := exec.Command(doltPath, "sql", "-q", query, "-r", "csv")
	sqlCmd.Dir = cloneDir
	output, err := sqlCmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("query failed: %s", string(exitErr.Stderr))
		}
		return fmt.Errorf("running query: %w", err)
	}

	rows := parseCSV(string(output))
	if len(rows) <= 1 {
		fmt.Println("No wanted items found matching your filters.")
		return nil
	}

	// Build styled table
	tbl := style.NewTable(
		style.Column{Name: "ID", Width: 12},
		style.Column{Name: "TITLE", Width: 40},
		style.Column{Name: "PROJECT", Width: 12},
		style.Column{Name: "TYPE", Width: 10},
		style.Column{Name: "PRI", Width: 4, Align: style.AlignRight},
		style.Column{Name: "POSTED BY", Width: 16},
		style.Column{Name: "STATUS", Width: 10},
		style.Column{Name: "EFFORT", Width: 8},
	)

	for _, row := range rows[1:] { // skip header
		if len(row) < 8 {
			continue
		}
		// Format priority as human-readable
		pri := formatPriority(row[4])
		tbl.AddRow(row[0], row[1], row[2], row[3], pri, row[5], row[6], row[7])
	}

	fmt.Printf("Wanted items (%d):\n\n", len(rows)-1)
	fmt.Print(tbl.Render())

	return nil
}

// parseCSV is a simple CSV parser for Dolt output.
// Handles quoted fields containing commas.
func parseCSV(data string) [][]string {
	var rows [][]string
	for _, line := range strings.Split(strings.TrimSpace(data), "\n") {
		if line == "" {
			continue
		}
		rows = append(rows, parseCSVLine(line))
	}
	return rows
}

// parseCSVLine parses a single CSV line, handling quoted fields.
func parseCSVLine(line string) []string {
	var fields []string
	var field strings.Builder
	inQuote := false

	for i := 0; i < len(line); i++ {
		ch := line[i]
		switch {
		case ch == '"' && !inQuote:
			inQuote = true
		case ch == '"' && inQuote:
			// Check for escaped quote ("")
			if i+1 < len(line) && line[i+1] == '"' {
				field.WriteByte('"')
				i++
			} else {
				inQuote = false
			}
		case ch == ',' && !inQuote:
			fields = append(fields, field.String())
			field.Reset()
		default:
			field.WriteByte(ch)
		}
	}
	fields = append(fields, field.String())
	return fields
}

// formatPriority converts numeric priority to display string.
func formatPriority(pri string) string {
	switch pri {
	case "0":
		return "P0"
	case "1":
		return "P1"
	case "2":
		return "P2"
	case "3":
		return "P3"
	case "4":
		return "P4"
	default:
		return pri
	}
}

// runWLSync pulls upstream changes into the local wl-commons fork.
func runWLSync(cmd *cobra.Command, args []string) error {
	townRoot, err := workspace.FindFromCwdOrError()
	if err != nil {
		return fmt.Errorf("not in a Gas Town workspace: %w", err)
	}

	// Check dolt is available
	doltPath, err := exec.LookPath("dolt")
	if err != nil {
		return fmt.Errorf("dolt not found in PATH — install from https://docs.dolthub.com/introduction/installation")
	}

	// Look for local wl-commons fork in standard locations
	forkDir := findWLCommonsFork(townRoot)
	if forkDir == "" {
		return fmt.Errorf("no local wl-commons fork found\n\nTo create one:\n  dolt clone <your-org>/wl-commons\n  cd wl-commons\n  dolt remote add upstream hop/wl-commons")
	}

	fmt.Printf("Local fork: %s\n", style.Dim.Render(forkDir))

	if wlSyncDryRun {
		fmt.Printf("\n%s Dry run — checking upstream for changes...\n", style.Bold.Render("~"))
	} else {
		fmt.Printf("\nPulling from upstream (%s/%s)...\n", wlCommonsOrg, wlCommonsDB)
	}

	// Run dolt pull from upstream
	pullArgs := []string{"pull", "upstream", "main"}
	if wlSyncDryRun {
		// In dry-run mode, just fetch and show diff
		fetchCmd := exec.Command(doltPath, "fetch", "upstream")
		fetchCmd.Dir = forkDir
		fetchCmd.Stderr = os.Stderr
		if err := fetchCmd.Run(); err != nil {
			return fmt.Errorf("fetching upstream: %w", err)
		}

		// Show diff summary
		diffCmd := exec.Command(doltPath, "diff", "--stat", "HEAD", "upstream/main")
		diffCmd.Dir = forkDir
		diffCmd.Stdout = os.Stdout
		diffCmd.Stderr = os.Stderr
		if err := diffCmd.Run(); err != nil {
			// No diff or error — either way report
			fmt.Printf("%s Already up to date.\n", style.Bold.Render("✓"))
		}
		return nil
	}

	pullCmd := exec.Command(doltPath, pullArgs[0], pullArgs[1], pullArgs[2])
	pullCmd.Dir = forkDir
	pullCmd.Stdout = os.Stdout
	pullCmd.Stderr = os.Stderr
	if err := pullCmd.Run(); err != nil {
		return fmt.Errorf("pulling from upstream: %w", err)
	}

	fmt.Printf("\n%s Synced with upstream\n", style.Bold.Render("✓"))

	// Show summary of what changed
	fmt.Println("\nChecking for new items...")
	summaryQuery := `SELECT
		(SELECT COUNT(*) FROM wanted WHERE status = 'open') AS open_wanted,
		(SELECT COUNT(*) FROM wanted) AS total_wanted,
		(SELECT COUNT(*) FROM completions) AS total_completions,
		(SELECT COUNT(*) FROM stamps) AS total_stamps`

	summaryCmd := exec.Command(doltPath, "sql", "-q", summaryQuery, "-r", "csv")
	summaryCmd.Dir = forkDir
	out, err := summaryCmd.Output()
	if err == nil {
		rows := parseCSV(string(out))
		if len(rows) >= 2 && len(rows[1]) >= 4 {
			r := rows[1]
			fmt.Printf("\n  Open wanted:       %s\n", r[0])
			fmt.Printf("  Total wanted:      %s\n", r[1])
			fmt.Printf("  Total completions: %s\n", r[2])
			fmt.Printf("  Total stamps:      %s\n", r[3])
		}
	}

	return nil
}

// findWLCommonsFork looks for a local wl-commons Dolt database in standard locations.
func findWLCommonsFork(townRoot string) string {
	// Check common locations relative to town root
	candidates := []string{
		filepath.Join(townRoot, "wl-commons"),
		filepath.Join(townRoot, "..", "wl-commons"),
		filepath.Join(os.Getenv("HOME"), "wl-commons"),
	}

	for _, dir := range candidates {
		doltDir := filepath.Join(dir, ".dolt")
		if info, err := os.Stat(doltDir); err == nil && info.IsDir() {
			return dir
		}
	}

	return ""
}
