package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "archivist",
	Short: "A workflow-first ADR management tool",
	Long: `Archivist is a drop-in replacement for adr-tools that adds a richer CLI
and an interactive TUI for managing Architecture Decision Records.

It works directly on existing adr-tools repositories without requiring
an import, migration, or file layout conversion.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	return rootCmd.Execute()
}
