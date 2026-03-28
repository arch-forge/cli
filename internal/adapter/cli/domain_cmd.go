package cli

import (
	"fmt"

	"github.com/arch-forge/cli/internal/app"
	"github.com/spf13/cobra"
)

func newDomainCmd(uc *app.DomainAddUseCase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "domain",
		Short: "Manage bounded-context domains in your project",
	}
	cmd.AddCommand(newDomainAddCmd(uc))
	return cmd
}

func newDomainAddCmd(uc *app.DomainAddUseCase) *cobra.Command {
	var (
		projectDir string
		dryRun     bool
	)
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a new bounded-context domain to an existing project",
		Long: `Scaffolds a new domain module following your project's architecture.

The generated structure depends on the architecture declared in archforge.yaml:
  hexagonal/modular → internal/{name}/{domain,ports,application,adapters}
  clean/modular     → internal/{name}/{domain,usecase,ports,adapters}
  ddd               → internal/{name}/{domain,application,infrastructure}

Examples:
  arch_forge domain add payment
  arch_forge domain add order --dry-run
  arch_forge domain add notification --project-dir ./myapp`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if projectDir == "" {
				projectDir = "."
			}
			opts := app.DomainAddOptions{
				ProjectDir: projectDir,
				Name:       args[0],
				DryRun:     dryRun,
			}
			if err := uc.Execute(cmd.Context(), opts); err != nil {
				return err
			}
			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "Dry run: domain %q would be added.\n", args[0])
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Domain %q added successfully.\n", args[0])
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&projectDir, "project-dir", "", "Project directory (defaults to current directory)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without writing to disk")
	return cmd
}
