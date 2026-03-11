package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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

		pattern, err := regexp.Compile("(?i)" + args[0])
		if err != nil {
			return fmt.Errorf("invalid pattern: %w", err)
		}

		files, err := repo.ListFiles()
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		for _, f := range files {
			absPath := f
			if !filepath.IsAbs(f) {
				absPath = filepath.Join(cwd, f)
			}

			data, err := os.ReadFile(absPath)
			if err != nil {
				continue
			}

			lines := strings.Split(string(data), "\n")
			var matches []string
			for i, line := range lines {
				if pattern.MatchString(line) {
					matches = append(matches, fmt.Sprintf("  %d: %s", i+1, line))
				}
			}

			if len(matches) > 0 {
				fmt.Fprintln(out, f)
				for _, m := range matches {
					fmt.Fprintln(out, m)
				}
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
