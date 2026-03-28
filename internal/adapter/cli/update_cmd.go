package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/arch-forge/cli/internal/app"
	"github.com/arch-forge/cli/internal/domain"
	"github.com/spf13/cobra"
)

// newUpdateCmd creates the `update` subcommand.
func newUpdateCmd(uc *app.UpdateUseCase) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update arch_forge to the latest version",
		Long: `update checks GitHub releases for a newer version of arch_forge and, if one
is found, downloads and atomically replaces the current binary.

Use --force to bypass the "already up to date" check and replace the binary
even when the running version matches the latest release.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := uc.Execute(app.UpdateOptions{Force: force})
			if err != nil {
				if errors.Is(err, domain.ErrDevVersion) {
					fmt.Println("Cannot update a dev build. Install a tagged release first:")
					fmt.Println("  brew install archforge/tap/arch-forge")
					return nil
				}

				var pathErr *os.PathError
				if errors.As(err, &pathErr) {
					fmt.Fprintf(os.Stderr, "Permission denied updating binary. Try: sudo arch_forge update\n")
					return err
				}

				return fmt.Errorf("update: %w", err)
			}

			if result.AlreadyUpToDate {
				fmt.Printf("arch_forge %s is already up to date.\n", Version)
				return nil
			}

			fmt.Printf("Updated arch_forge %s → %s\n", result.PreviousVersion, result.NewVersion)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip the \"already up to date\" check and replace the binary unconditionally")

	return cmd
}
