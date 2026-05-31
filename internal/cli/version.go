package cli

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var Version = "dev"

func currentVersion(moduleVersion string) string {
	if Version != "" && Version != "dev" {
		return Version
	}
	if moduleVersion != "" && moduleVersion != "(devel)" {
		return moduleVersion
	}
	if Version == "" {
		return "dev"
	}
	return Version
}

func buildInfoModuleVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	return info.Main.Version
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the archivist version",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(cmd.OutOrStdout(), "archivist "+currentVersion(buildInfoModuleVersion()))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
