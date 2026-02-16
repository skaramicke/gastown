package cmd

import (
	"github.com/spf13/cobra"
)

var wlCmd = &cobra.Command{
	Use:     "wl",
	GroupID: GroupWork,
	Short:   "Wasteland federation commands",
	RunE:    requireSubcommand,
	Long: `Wasteland federation commands for interacting with the shared commons.

The Wasteland is a federation protocol built on DoltHub. Towns collaborate
through a shared commons database (wl-commons) containing wanted items,
completions, and reputation data.

Commands:
  post    Create a wanted item on the commons
  browse  Browse wanted items (clone-then-discard)
  sync    Pull upstream changes into local fork`,
}

func init() {
	rootCmd.AddCommand(wlCmd)
}
