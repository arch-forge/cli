package cli

import (
	"fmt"
	"os"

	"github.com/arch-forge/cli/internal/adapter/graph"
	"github.com/arch-forge/cli/internal/app"
	"github.com/spf13/cobra"
)

// newGraphCmd creates the `graph` subcommand.
func newGraphCmd(uc *app.GraphUseCase) *cobra.Command {
	var projectDir string
	var format string
	var includeExternal bool
	var output string

	cmd := &cobra.Command{
		Use:   "graph",
		Short: "Generate a dependency diagram for a Go project",
		Long:  `graph analyzes a Go project's import graph and outputs a Mermaid or DOT dependency diagram.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := app.GraphOptions{
				ProjectDir:      projectDir,
				Format:          format,
				IncludeExternal: includeExternal,
			}

			g, err := uc.Execute(opts)
			if err != nil {
				return fmt.Errorf("graph: %w", err)
			}

			// Determine output destination.
			w := cmd.OutOrStdout()
			if output != "" {
				f, openErr := os.Create(output)
				if openErr != nil {
					return fmt.Errorf("graph: open output file: %w", openErr)
				}
				defer f.Close()
				w = f
			}

			// Select renderer based on format.
			switch format {
			case "dot":
				if renderErr := graph.RenderDOT(w, g); renderErr != nil {
					return fmt.Errorf("graph: render DOT: %w", renderErr)
				}
			default:
				// Default to mermaid.
				if renderErr := graph.RenderMermaid(w, g); renderErr != nil {
					return fmt.Errorf("graph: render Mermaid: %w", renderErr)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&projectDir, "project-dir", ".", "Path to the project root directory")
	cmd.Flags().StringVar(&format, "format", "mermaid", "Output format: mermaid or dot")
	cmd.Flags().BoolVar(&includeExternal, "include-external", false, "Include third-party packages as nodes")
	cmd.Flags().StringVar(&output, "output", "", "Write output to file instead of stdout")

	return cmd
}
