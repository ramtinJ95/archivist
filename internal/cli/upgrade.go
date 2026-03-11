package cli

import (
	"fmt"
	"os"

	"github.com/ramtinJ95/archivist/internal/adrlog"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade-repository",
	Short: "Upgrade ADR repository to latest format",
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

		count, err := repo.UpgradeRepository()
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Upgraded %d file(s)\n", count)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
