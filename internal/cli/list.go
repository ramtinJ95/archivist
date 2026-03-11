package cli

import (
	"fmt"
	"os"

	"github.com/ramtinJ95/archivist/internal/adrlog"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all ADR files",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		repo, err := adrlog.OpenRepository(cwd)
		if err != nil {
			return err
		}

		files, err := repo.ListFiles()
		if err != nil {
			return err
		}

		for _, f := range files {
			fmt.Fprintln(cmd.OutOrStdout(), f)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
