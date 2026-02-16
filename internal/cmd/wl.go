package cmd

import (
	"github.com/spf13/cobra"
)

var wlCmd = &cobra.Command{
	Use:     "wl",
	GroupID: GroupWork,
	Short:   "Wasteland federation commands (wanted board, reputation)",
	RunE:    requireSubcommand,
	Long: `Wasteland federation commands for the shared wanted board.

The Wasteland is a federation protocol for posting work, claiming tasks,
and building reputation across Gas Towns via DoltHub.

Commands:
  gt wl post    Post a new wanted item to the commons`,
}

func init() {
	rootCmd.AddCommand(wlCmd)
}
