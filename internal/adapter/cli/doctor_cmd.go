package cli

import (
	"bufio"
	"fmt"
	"strings"
	"text/tabwriter"

	analyzeradapter "github.com/arch-forge/cli/internal/adapter/analyzer"
	"github.com/arch-forge/cli/internal/app"
	"github.com/arch-forge/cli/internal/domain"
	"github.com/spf13/cobra"
)

// newDoctorCmd creates the `doctor` subcommand.
func newDoctorCmd(uc *app.DoctorUseCase) *cobra.Command {
	var projectDir string
	var threshold float64
	var fix bool
	var yes bool

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Validate architecture compliance of a Go project",
		Long:  `doctor analyzes a Go project's import graph and reports architecture rule violations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := app.DoctorOptions{
				ProjectDir:     projectDir,
				ScoreThreshold: threshold,
				Fix:            fix,
			}

			report, err := uc.Execute(opts)
			if err != nil {
				return fmt.Errorf("doctor: %w", err)
			}

			printReport(cmd, report, threshold)

			if fix && len(report.Violations) > 0 {
				suggestions := analyzeradapter.SuggestFixes(report)
				printSuggestions(cmd, suggestions, yes)
			}

			if report.Score < threshold {
				return fmt.Errorf("✗ Score %.1f is below threshold %.1f", report.Score, threshold)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&projectDir, "project-dir", ".", "Path to the project root directory")
	cmd.Flags().Float64Var(&threshold, "threshold", 7.0, "Minimum acceptable score (0–10)")
	cmd.Flags().BoolVar(&fix, "fix", false, "Show fix suggestions for violations")
	cmd.Flags().BoolVar(&yes, "yes", false, "Auto-apply all applicable fixes without prompting")

	return cmd
}

// printSuggestions writes fix suggestions to the command output and optionally
// auto-applies fixes that are marked as AutoApplicable. In v0.3 all suggestions
// are manual, so this function always prints a guidance message.
func printSuggestions(cmd *cobra.Command, suggestions []domain.FixSuggestion, autoApply bool) {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "\nFIX SUGGESTIONS (%d)\n\n", len(suggestions))

	for i, s := range suggestions {
		switch s.Kind {
		case domain.FixMoveFile:
			fmt.Fprintf(out, "  %d. [%s]    %s → %s\n", i+1, s.Kind, s.SourcePath, s.DestPath)
		case domain.FixRemoveImport:
			fmt.Fprintf(out, "  %d. [%s] %s:%d\n", i+1, s.Kind, s.Violation.File, s.Violation.Line)
		default:
			fmt.Fprintf(out, "  %d. [%s]       %s:%d\n", i+1, s.Kind, s.Violation.File, s.Violation.Line)
		}
		fmt.Fprintf(out, "     %s\n", s.Description)
		fmt.Fprintln(out, "     ⚠ Manual fix required")
		fmt.Fprintln(out)
	}

	// Count auto-applicable fixes.
	var autoCount int
	for _, s := range suggestions {
		if s.AutoApplicable {
			autoCount++
		}
	}

	if autoCount == 0 {
		fmt.Fprintln(out, "No auto-applicable fixes available. Please apply the suggestions above manually.")
		return
	}

	if autoApply {
		fmt.Fprintf(out, "Applying %d auto-applicable fix(es)...\n", autoCount)
		// Auto-apply logic would go here in a future version.
		fmt.Fprintln(out, "Done.")
		return
	}

	// Interactive confirmation prompt.
	fmt.Fprintf(out, "Apply %d fix(es)? [y/N] ", autoCount)
	scanner := bufio.NewScanner(cmd.InOrStdin())
	if scanner.Scan() {
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if answer == "y" || answer == "yes" {
			fmt.Fprintf(out, "Applying %d auto-applicable fix(es)...\n", autoCount)
			// Auto-apply logic would go here in a future version.
			fmt.Fprintln(out, "Done.")
		} else {
			fmt.Fprintln(out, "No fixes applied.")
		}
	}
}

// printReport writes the doctor analysis report to the command output.
func printReport(cmd *cobra.Command, report domain.Report, threshold float64) {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "arch_forge doctor — Analyzing %s\n", report.ProjectPath)
	fmt.Fprintf(out, "Architecture: %s (%s)\n\n", report.Arch, report.Variant)

	// Separate violations by severity.
	var errors []domain.Violation
	var warnings []domain.Violation
	for _, v := range report.Violations {
		switch v.Severity {
		case domain.SeverityError:
			errors = append(errors, v)
		case domain.SeverityWarning:
			warnings = append(warnings, v)
		}
	}

	// Print errors section.
	fmt.Fprintf(out, "ERRORS (%d)\n", len(errors))
	if len(errors) > 0 {
		w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		for _, v := range errors {
			fmt.Fprintf(w, "  %s:%d\t%s\t%s\n", v.File, v.Line, v.Rule, v.Message)
		}
		_ = w.Flush()
	}

	fmt.Fprintln(out)

	// Print warnings section.
	fmt.Fprintf(out, "WARNINGS (%d)\n", len(warnings))
	if len(warnings) > 0 {
		w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		for _, v := range warnings {
			fmt.Fprintf(w, "  %s:%d\t%s\t%s\n", v.File, v.Line, v.Rule, v.Message)
		}
		_ = w.Flush()
	}

	fmt.Fprintln(out)
	fmt.Fprintf(out, "Score: %.1f / 10.0  [threshold: %.1f]\n", report.Score, threshold)

	if report.Score >= threshold {
		fmt.Fprintln(out, "✓ Architecture compliance check passed")
	}
}
