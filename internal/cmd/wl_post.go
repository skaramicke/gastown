package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/steveyegge/gastown/internal/style"
	"github.com/steveyegge/gastown/internal/wasteland"
	"github.com/steveyegge/gastown/internal/workspace"
)

var (
	wlPostTitle       string
	wlPostDescription string
	wlPostProject     string
	wlPostType        string
	wlPostPriority    int
	wlPostEffort      string
	wlPostTags        string
	wlPostCommons     string
	wlPostDryRun      bool
)

var wlPostCmd = &cobra.Command{
	Use:   "post",
	Short: "Create a wanted item on the commons",
	Long: `Create a wanted item on the shared Wasteland commons.

This command clones the commons database, inserts a new wanted item,
commits the change, and pushes it back to DoltHub.

Phase 1 (wild-west mode): direct write to main branch.

Examples:
  gt wl post --title "Fix auth bug" --project gastown --type bug
  gt wl post --title "Add pagination" --project gastown --type feature --priority 1 --effort large
  gt wl post --title "Update docs" --tags "docs,federation" --effort small`,
	RunE: runWlPost,
}

func init() {
	wlPostCmd.Flags().StringVar(&wlPostTitle, "title", "", "Title of the wanted item (required)")
	wlPostCmd.Flags().StringVar(&wlPostDescription, "description", "", "Detailed description")
	wlPostCmd.Flags().StringVarP(&wlPostProject, "project", "p", "", "Project name (e.g., gastown, beads)")
	wlPostCmd.Flags().StringVarP(&wlPostType, "type", "t", "feature", "Type: feature, bug, design, rfc, docs")
	wlPostCmd.Flags().IntVar(&wlPostPriority, "priority", 2, "Priority: 0=critical, 1=high, 2=medium, 3=low, 4=backlog")
	wlPostCmd.Flags().StringVar(&wlPostEffort, "effort", "medium", "Effort level: trivial, small, medium, large, epic")
	wlPostCmd.Flags().StringVar(&wlPostTags, "tags", "", "Comma-separated tags (e.g., \"go,auth,federation\")")
	wlPostCmd.Flags().StringVar(&wlPostCommons, "commons", wasteland.UpstreamCommons, "Commons database (org/repo)")
	wlPostCmd.Flags().BoolVar(&wlPostDryRun, "dry-run", false, "Show what would be posted without writing")

	_ = wlPostCmd.MarkFlagRequired("title")

	wlCmd.AddCommand(wlPostCmd)
}

func runWlPost(cmd *cobra.Command, args []string) error {
	// Resolve town handle for posted_by
	townRoot, err := workspace.FindFromCwdOrError()
	if err != nil {
		return fmt.Errorf("not in a Gas Town workspace: %w", err)
	}

	townHandle, err := wasteland.GetTownHandle(townRoot)
	if err != nil {
		return fmt.Errorf("resolving town handle: %w", err)
	}

	// Validate inputs
	validTypes := map[string]bool{"feature": true, "bug": true, "design": true, "rfc": true, "docs": true}
	if !validTypes[wlPostType] {
		return fmt.Errorf("invalid type %q: must be one of feature, bug, design, rfc, docs", wlPostType)
	}

	validEfforts := map[string]bool{"trivial": true, "small": true, "medium": true, "large": true, "epic": true}
	if !validEfforts[wlPostEffort] {
		return fmt.Errorf("invalid effort %q: must be one of trivial, small, medium, large, epic", wlPostEffort)
	}

	if wlPostPriority < 0 || wlPostPriority > 4 {
		return fmt.Errorf("invalid priority %d: must be 0-4", wlPostPriority)
	}

	// Generate wanted item ID
	id := wasteland.GenerateWantedID(wlPostTitle)

	// Build tags JSON
	tagsJSON := "NULL"
	if wlPostTags != "" {
		tagList := strings.Split(wlPostTags, ",")
		for i := range tagList {
			tagList[i] = strings.TrimSpace(tagList[i])
		}
		tagsBytes, err := json.Marshal(tagList)
		if err != nil {
			return fmt.Errorf("encoding tags: %w", err)
		}
		tagsJSON = fmt.Sprintf("'%s'", wasteland.EscapeSQL(string(tagsBytes)))
	}

	now := time.Now().UTC().Format("2006-01-02 15:04:05")

	// Build the INSERT statement
	descValue := "NULL"
	if wlPostDescription != "" {
		descValue = fmt.Sprintf("'%s'", wasteland.EscapeSQL(wlPostDescription))
	}

	projectValue := "NULL"
	if wlPostProject != "" {
		projectValue = fmt.Sprintf("'%s'", wasteland.EscapeSQL(wlPostProject))
	}

	insertSQL := fmt.Sprintf(
		`INSERT INTO wanted (id, title, description, project, type, priority, tags, posted_by, status, effort_level, created_at, updated_at) VALUES ('%s', '%s', %s, %s, '%s', %d, %s, '%s', 'open', '%s', '%s', '%s');`,
		wasteland.EscapeSQL(id),
		wasteland.EscapeSQL(wlPostTitle),
		descValue,
		projectValue,
		wasteland.EscapeSQL(wlPostType),
		wlPostPriority,
		tagsJSON,
		wasteland.EscapeSQL(townHandle),
		wasteland.EscapeSQL(wlPostEffort),
		now,
		now,
	)

	// Dry-run: show what would happen
	if wlPostDryRun {
		fmt.Printf("%s Dry run — no changes will be made\n\n", style.Bold.Render("!"))
		fmt.Printf("  Commons:   %s\n", wlPostCommons)
		fmt.Printf("  ID:        %s\n", id)
		fmt.Printf("  Title:     %s\n", wlPostTitle)
		if wlPostDescription != "" {
			fmt.Printf("  Desc:      %s\n", wlPostDescription)
		}
		if wlPostProject != "" {
			fmt.Printf("  Project:   %s\n", wlPostProject)
		}
		fmt.Printf("  Type:      %s\n", wlPostType)
		fmt.Printf("  Priority:  %d\n", wlPostPriority)
		fmt.Printf("  Effort:    %s\n", wlPostEffort)
		if wlPostTags != "" {
			fmt.Printf("  Tags:      %s\n", wlPostTags)
		}
		fmt.Printf("  Posted by: %s\n", townHandle)
		fmt.Printf("\n  SQL:\n    %s\n", insertSQL)
		return nil
	}

	// Step 1: Clone the commons
	fmt.Printf("Cloning %s...\n", wlPostCommons)
	cloneDir, err := wasteland.CloneCommons(wlPostCommons)
	if err != nil {
		return fmt.Errorf("cloning commons: %w", err)
	}
	defer func() {
		fmt.Printf("Cleaning up temporary clone...\n")
		// cloneDir is tmpDir/wl-commons; remove the parent tmpDir to fully clean up
		_ = os.RemoveAll(filepath.Dir(cloneDir))
	}()
	fmt.Printf("%s Cloned to temporary directory\n", style.Bold.Render("✓"))

	// Step 2: INSERT the wanted item
	fmt.Printf("Inserting wanted item %s...\n", id)
	if _, err := wasteland.DoltSQL(cloneDir, insertSQL); err != nil {
		return fmt.Errorf("inserting wanted item: %w", err)
	}
	fmt.Printf("%s Inserted wanted item\n", style.Bold.Render("✓"))

	// Step 3: Commit
	commitMsg := fmt.Sprintf("wanted: %s — %s (posted by %s)", id, wlPostTitle, townHandle)
	if err := wasteland.DoltCommit(cloneDir, commitMsg); err != nil {
		return fmt.Errorf("committing: %w", err)
	}
	fmt.Printf("%s Committed\n", style.Bold.Render("✓"))

	// Step 4: Push to origin
	fmt.Printf("Pushing to %s...\n", wlPostCommons)
	if err := wasteland.DoltPush(cloneDir); err != nil {
		return fmt.Errorf("pushing to commons: %w", err)
	}
	fmt.Printf("%s Pushed to commons\n", style.Bold.Render("✓"))

	// Summary
	fmt.Printf("\n%s Posted wanted item:\n", style.Bold.Render("✓"))
	fmt.Printf("  ID:       %s\n", style.Bold.Render(id))
	fmt.Printf("  Title:    %s\n", wlPostTitle)
	if wlPostProject != "" {
		fmt.Printf("  Project:  %s\n", wlPostProject)
	}
	fmt.Printf("  Type:     %s\n", wlPostType)
	fmt.Printf("  Priority: %d\n", wlPostPriority)
	fmt.Printf("  Effort:   %s\n", wlPostEffort)
	fmt.Printf("  Posted:   %s\n", townHandle)

	return nil
}
