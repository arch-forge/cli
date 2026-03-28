package cli

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/arch-forge/cli/internal/app"
	"github.com/arch-forge/cli/internal/domain"
	"github.com/spf13/cobra"
)

// newModuleCmd creates the `module` command with subcommands.
func newModuleCmd(uc *app.ModuleUseCase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "module",
		Short: "Create and manage custom local modules",
	}
	cmd.AddCommand(newModuleCreateCmd(uc))
	cmd.AddCommand(newModuleDevCmd(uc))
	cmd.AddCommand(newModuleValidateCmd(uc))
	return cmd
}

// newModuleCreateCmd returns the `module create <name>` subcommand.
func newModuleCreateCmd(uc *app.ModuleUseCase) *cobra.Command {
	var (
		category string
		dir      string
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Scaffold a new local module directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			opts := app.ModuleCreateOptions{
				Name:     name,
				Category: category,
				Dir:      dir,
			}

			if err := uc.Create(opts); err != nil {
				return fmt.Errorf("module create: %w", err)
			}

			resolvedDir := dir
			if resolvedDir == "" {
				resolvedDir = "modules/"
			}
			modulePath := filepath.Join(resolvedDir, name)
			fmt.Fprintf(cmd.OutOrStdout(), "✓ Module %q scaffolded at %s\n", name, modulePath)
			return nil
		},
	}

	cmd.Flags().StringVar(&category, "category", "custom", "Module category")
	cmd.Flags().StringVar(&dir, "dir", "modules/", "Base directory for the module")

	return cmd
}

// newModuleValidateCmd returns the `module validate <name|path>` subcommand.
func newModuleValidateCmd(uc *app.ModuleUseCase) *cobra.Command {
	var dir string

	cmd := &cobra.Command{
		Use:   "validate <name|path>",
		Short: "Validate a module's manifest, templates, and path resolution",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg := args[0]

			var moduleDir string
			if strings.Contains(arg, "/") || strings.Contains(arg, string(os.PathSeparator)) {
				moduleDir = arg
			} else {
				moduleDir = filepath.Join(dir, arg)
			}

			opts := app.ModuleValidateOptions{
				ModuleDir: moduleDir,
			}

			result, err := uc.Validate(opts)
			if err != nil {
				return fmt.Errorf("module validate: %w", err)
			}

			printValidationResult(cmd, result)

			if result.Status == domain.LocalModuleStatusInvalid {
				return fmt.Errorf("module %q failed validation", result.ModuleName)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&dir, "dir", "modules/", "Base directory where modules are stored (used when resolving by name)")

	return cmd
}

// newModuleDevCmd returns the `module dev <name>` subcommand.
func newModuleDevCmd(uc *app.ModuleUseCase) *cobra.Command {
	var (
		dir      string
		interval time.Duration
	)

	cmd := &cobra.Command{
		Use:   "dev <name>",
		Short: "Watch mode for template development (re-validates on interval)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			moduleDir := filepath.Join(dir, name)

			fmt.Fprintf(cmd.OutOrStdout(), "Watching %s for changes (Ctrl+C to stop)...\n", moduleDir)

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			defer signal.Stop(sigCh)

			opts := app.ModuleValidateOptions{ModuleDir: moduleDir}

			for {
				result, err := uc.Validate(opts)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error: %s\n", err.Error())
				} else {
					printValidationResult(cmd, result)
				}

				select {
				case <-sigCh:
					fmt.Fprintln(cmd.OutOrStdout(), "Stopping watch mode.")
					return nil
				case <-time.After(interval):
				}
			}
		},
	}

	cmd.Flags().StringVar(&dir, "dir", "modules/", "Base directory where modules are stored")
	cmd.Flags().DurationVar(&interval, "interval", 2*time.Second, "Interval between validation runs")

	return cmd
}

// printValidationResult writes a formatted validation report to the command's output.
func printValidationResult(cmd *cobra.Command, result domain.LocalModuleValidation) {
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "\nValidating module %q at %s\n\n", result.ModuleName, result.ModuleDir)

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)

	errorCount := 0
	warnCount := 0

	for _, issue := range result.Issues {
		switch issue.Kind {
		case "error":
			errorCount++
			fmt.Fprintf(w, "  ✗ %s\n", issue.Message)
		case "warning":
			warnCount++
			fmt.Fprintf(w, "  ⚠ %s\n", issue.Message)
		}
	}

	if len(result.Issues) == 0 {
		fmt.Fprintf(w, "  ✓ all checks passed\n")
	}

	_ = w.Flush()

	statusLabel := strings.ToUpper(string(result.Status))
	fmt.Fprintf(out, "\nStatus: %s (%d warning(s), %d error(s))\n", statusLabel, warnCount, errorCount)
}
