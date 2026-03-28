package cli

import (
	"fmt"
	"os"

	"github.com/arch-forge/cli/internal/app"
	"github.com/spf13/cobra"
)

// newInspectCmd creates the `inspect` subcommand.
func newInspectCmd(uc *app.InspectUseCase) *cobra.Command {
	var projectDir string
	var depth int
	var format string

	cmd := &cobra.Command{
		Use:   "inspect",
		Short: "Show project structure overview",
		Long:  `inspect reads archforge.yaml and prints an annotated directory tree of the project.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := app.InspectOptions{
				ProjectDir: projectDir,
				MaxDepth:   depth,
			}

			summary, err := uc.Execute(opts)
			if err != nil {
				return fmt.Errorf("inspect: %w", err)
			}

			switch format {
			case "json":
				return RenderJSON(cmd.OutOrStdout(), summary)
			default:
				return RenderTree(cmd.OutOrStdout(), summary, isTerminal(os.Stdout))
			}
		},
	}

	cmd.Flags().StringVar(&projectDir, "project-dir", "", "Path to the project root directory (default: current directory)")
	cmd.Flags().IntVar(&depth, "depth", 3, "Maximum directory depth to traverse (0 = unlimited)")
	cmd.Flags().StringVar(&format, "format", "tree", "Output format: tree or json")

	return cmd
}
