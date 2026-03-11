package cli

import (
	"fmt"
	"os"

	"github.com/ramtinJ95/archivist/internal/adrlog"
	"github.com/ramtinJ95/archivist/internal/tui"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive terminal UI for browsing ADRs",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting working directory: %w", err)
		}

		repo, err := adrlog.OpenRepository(cwd)
		if err != nil {
			return fmt.Errorf("opening repository: %w", err)
		}

		return tui.Run(repo)
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
