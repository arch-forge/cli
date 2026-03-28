package cli

import (
	"fmt"

	"github.com/arch-forge/cli/internal/app"
	"github.com/spf13/cobra"
)

func newAddCmd(uc *app.AddUseCase) *cobra.Command {
	var (
		projectDir string
		dryRun     bool
	)

	cmd := &cobra.Command{
		Use:   "add <module> [module...]",
		Short: "Add one or more modules to an existing project",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if projectDir == "" {
				projectDir = "."
			}

			opts := app.AddOptions{
				ProjectDir:    projectDir,
				Modules:       args,
				ModuleOptions: make(map[string]map[string]any),
				DryRun:        dryRun,
			}

			if err := uc.Execute(cmd.Context(), opts); err != nil {
				return fmt.Errorf("add: %w", err)
			}

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "Dry run complete. No files written.\n")
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Module(s) %v added successfully.\n", args)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&projectDir, "project-dir", "", "Project directory (defaults to current directory)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without writing to disk")

	// Register dynamic completion for positional module name arguments.
	cmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"api\tHTTP API server with chi router",
			"database\tPostgreSQL connection and pool",
			"logging\tStructured logging with slog",
			"docker\tMulti-stage Dockerfile and docker-compose",
			"makefile\tStandard Makefile targets",
			"auth\tJWT authentication middleware",
			"cache\tRedis cache client",
			"grpc\tgRPC server with interceptors",
			"crud\tCRUD scaffold for an entity",
		}, cobra.ShellCompDirectiveNoFileComp
	}

	return cmd
}
