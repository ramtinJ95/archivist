package cli

import (
	"fmt"
	"os"

	"github.com/ramtinJ95/archivist/internal/adrlog"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search PATTERN",
	Short: "Search ADR content with a case-insensitive regex",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		repo, err := adrlog.OpenRepository(cwd)
		if err != nil {
			return err
		}

		results, err := repo.Search(args[0])
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		for _, r := range results {
			fmt.Fprintln(out, r.Path)
			for _, m := range r.Matches {
				fmt.Fprintf(out, "  %d: %s\n", m.Line, m.Content)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
