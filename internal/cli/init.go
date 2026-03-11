package cli

import (
	"fmt"
	"os"

	"github.com/ramtinJ95/archivist/internal/adrlog"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [dir]",
	Short: "Initialize a new ADR repository",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		dir := ""
		if len(args) > 0 {
			dir = args[0]
		}

		path, err := adrlog.InitRepository(cwd, dir)
		if err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
