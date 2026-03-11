package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ramtinJ95/archivist/internal/adrlog"
	"github.com/ramtinJ95/archivist/internal/editor"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new TITLE...",
	Short: "Create a new ADR",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		repo, err := adrlog.OpenRepository(cwd)
		if err != nil {
			return err
		}

		supersedes, _ := cmd.Flags().GetStringSlice("supersedes")
		linkSpecs, _ := cmd.Flags().GetStringSlice("link")

		var links []adrlog.LinkSpec
		for _, spec := range linkSpecs {
			ls, err := adrlog.ParseLinkSpec(spec)
			if err != nil {
				return err
			}
			links = append(links, ls)
		}

		opts := adrlog.CreateOptions{
			Title:      strings.Join(args, " "),
			Supersedes: supersedes,
			Links:      links,
		}

		relPath, err := repo.CreateADR(opts)
		if err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), relPath)

		absPath := filepath.Join(cwd, relPath)
		return editor.LaunchEditor(absPath)
	},
}

func init() {
	newCmd.Flags().StringSliceP("supersedes", "s", nil, "ADR number(s) this supersedes")
	newCmd.Flags().StringSliceP("link", "l", nil, "Link spec as TARGET:LINK:REVERSE-LINK")
	rootCmd.AddCommand(newCmd)
}
