package cli

import (
	"fmt"
	"os"

	"github.com/ramtinJ95/archivist/internal/adrlog"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Check all ADRs for common issues",
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

		issues, err := repo.Validate()
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		for _, issue := range issues {
			fmt.Fprintf(out, "%s: [%s] %s\n", issue.Path, issue.Severity, issue.Message)
		}

		if len(issues) > 0 {
			return fmt.Errorf("%d validation issue(s) found", len(issues))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
