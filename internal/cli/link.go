package cli

import (
	"os"
	"path/filepath"

	"github.com/ramtinJ95/archivist/internal/adrlog"
	"github.com/spf13/cobra"
)

var linkCmd = &cobra.Command{
	Use:   "link SOURCE LINK TARGET REVERSE-LINK",
	Short: "Create a link between two ADRs",
	Args:  cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		repo, err := adrlog.OpenRepository(cwd)
		if err != nil {
			return err
		}

		sourcePath, err := repo.ResolveRef(args[0])
		if err != nil {
			return err
		}
		targetPath, err := repo.ResolveRef(args[2])
		if err != nil {
			return err
		}

		if !filepath.IsAbs(sourcePath) {
			sourcePath = filepath.Join(cwd, sourcePath)
		}
		if !filepath.IsAbs(targetPath) {
			targetPath = filepath.Join(cwd, targetPath)
		}

		return adrlog.AddLink(sourcePath, targetPath, args[1], args[3])
	},
}

func init() {
	rootCmd.AddCommand(linkCmd)
}
