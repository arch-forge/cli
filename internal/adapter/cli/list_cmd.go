package cli

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/arch-forge/cli/internal/app"
	"github.com/spf13/cobra"
)

func newListCmd(uc *app.ListUseCase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available architectures and modules",
	}

	cmd.AddCommand(newListArchsCmd(uc))
	cmd.AddCommand(newListModulesCmd(uc))
	cmd.AddCommand(newListPresetsCmd(uc))

	return cmd
}

func newListArchsCmd(uc *app.ListUseCase) *cobra.Command {
	return &cobra.Command{
		Use:   "archs",
		Short: "List all supported architecture patterns",
		RunE: func(cmd *cobra.Command, args []string) error {
			archs := uc.Architectures()
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tDISPLAY NAME\tDESCRIPTION")
			fmt.Fprintln(w, "----\t------------\t-----------")
			for _, a := range archs {
				fmt.Fprintf(w, "%s\t%s\t%s\n", a.Value, a.DisplayName, a.Description)
			}
			return w.Flush()
		},
	}
}

func newListModulesCmd(uc *app.ListUseCase) *cobra.Command {
	var category string

	cmd := &cobra.Command{
		Use:   "modules",
		Short: "List all available modules",
		RunE: func(cmd *cobra.Command, args []string) error {
			modules, err := uc.Modules(cmd.Context(), category)
			if err != nil {
				return fmt.Errorf("list modules: %w", err)
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tCATEGORY\tDESCRIPTION")
			fmt.Fprintln(w, "----\t--------\t-----------")
			for _, m := range modules {
				fmt.Fprintf(w, "%s\t%s\t%s\n", m.Name, m.Category, m.Description)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&category, "category", "", "Filter by category (core, infrastructure, observability, devops, testing, security)")

	return cmd
}

func newListPresetsCmd(uc *app.ListUseCase) *cobra.Command {
	return &cobra.Command{
		Use:   "presets",
		Short: "List all available presets",
		RunE: func(cmd *cobra.Command, args []string) error {
			presets := uc.Presets()
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tARCH\tVARIANT\tMODULES")
			fmt.Fprintln(w, "----\t----\t-------\t-------")
			for _, p := range presets {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					p.Name,
					p.Arch,
					p.Variant,
					strings.Join(p.Modules, ", "),
				)
			}
			return w.Flush()
		},
	}
}
