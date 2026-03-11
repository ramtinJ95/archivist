package cli

import (
	"fmt"
	"os"

	"github.com/ramtinJ95/archivist/internal/adrlog"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate documentation from ADRs",
}

var generateTOCCmd = &cobra.Command{
	Use:   "toc",
	Short: "Generate a table of contents",
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

		intro, _ := cmd.Flags().GetString("intro")
		outro, _ := cmd.Flags().GetString("outro")
		linkPrefix, _ := cmd.Flags().GetString("link-prefix")

		toc, err := repo.GenerateTOC(adrlog.TOCOptions{
			Intro:      intro,
			Outro:      outro,
			LinkPrefix: linkPrefix,
		})
		if err != nil {
			return err
		}

		fmt.Fprint(cmd.OutOrStdout(), toc)
		return nil
	},
}

var generateGraphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Generate a dependency graph in DOT format",
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

		linkPrefix, _ := cmd.Flags().GetString("link-prefix")
		extension, _ := cmd.Flags().GetString("extension")

		graph, err := repo.GenerateGraph(adrlog.GraphOptions{
			LinkPrefix:    linkPrefix,
			LinkExtension: extension,
		})
		if err != nil {
			return err
		}

		fmt.Fprint(cmd.OutOrStdout(), graph)
		return nil
	},
}

func init() {
	generateTOCCmd.Flags().StringP("intro", "i", "", "Introductory text")
	generateTOCCmd.Flags().StringP("outro", "o", "", "Closing text")
	generateTOCCmd.Flags().StringP("link-prefix", "p", "", "Prefix for links")

	generateGraphCmd.Flags().StringP("link-prefix", "p", "", "Prefix for links")
	generateGraphCmd.Flags().StringP("extension", "e", "", "File extension for links")

	generateCmd.AddCommand(generateTOCCmd)
	generateCmd.AddCommand(generateGraphCmd)
	rootCmd.AddCommand(generateCmd)
}
