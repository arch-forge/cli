// Package cli provides the command-line interface adapter for arch_forge.
package cli

import (
	"fmt"
	"os"

	"github.com/arch-forge/cli/internal/app"
	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags.
var Version = "dev"

// Execute wires up all subcommands and runs the root cobra command.
func Execute(initUC *app.InitUseCase, addUC *app.AddUseCase, listUC *app.ListUseCase, doctorUC *app.DoctorUseCase, inspectUC *app.InspectUseCase, graphUC *app.GraphUseCase, moduleUC *app.ModuleUseCase, updateUC *app.UpdateUseCase, domainUC *app.DomainAddUseCase) {
	rootCmd := buildRootCmd()
	rootCmd.AddCommand(newInitCmd(initUC))
	rootCmd.AddCommand(newAddCmd(addUC))
	rootCmd.AddCommand(newListCmd(listUC))
	rootCmd.AddCommand(newDoctorCmd(doctorUC))
	rootCmd.AddCommand(newInspectCmd(inspectUC))
	rootCmd.AddCommand(newGraphCmd(graphUC))
	rootCmd.AddCommand(newModuleCmd(moduleUC))
	rootCmd.AddCommand(newUpdateCmd(updateUC))
	rootCmd.AddCommand(newDomainCmd(domainUC))
	rootCmd.AddCommand(newCompletionCmd(rootCmd))

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func buildRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:           "arch_forge",
		Version:       Version,
		Short:         "Generate Go project structures based on proven architectural patterns",
		Long:          `arch_forge scaffolds Go projects following architectural patterns such as Hexagonal, Clean Architecture, DDD, and more.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}
