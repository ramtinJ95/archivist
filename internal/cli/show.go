package cli

import (
	"os"
	"path/filepath"

	"github.com/ramtinJ95/archivist/internal/adrlog"
	"github.com/ramtinJ95/archivist/internal/editor"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show REF",
	Short: "Display the full content of an ADR",
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

		path, err := repo.ResolveRef(args[0])
		if err != nil {
			return err
		}

		if !filepath.IsAbs(path) {
			path = filepath.Join(cwd, path)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return editor.LaunchPager(cmd.OutOrStdout(), string(data))
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}
