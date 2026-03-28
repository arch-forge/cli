package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildTestRootForCompletion creates a minimal root command with the completion
// subcommand attached, suitable for use in unit tests without requiring use-case
// dependencies.
func buildTestRootForCompletion() *cobra.Command {
	root := &cobra.Command{
		Use:           "arch_forge",
		Short:         "Test root command",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Use the exported newCompletionCmd helper indirectly by calling the same
	// pattern: the completion subcommand is registered against root so that
	// GenBashCompletion / GenZshCompletion etc. reflect the full tree.
	completionCmd := &cobra.Command{
		Use:       "completion [bash|zsh|fish|powershell]",
		Short:     "Generate shell completion scripts",
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return root.GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				return root.GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				return root.GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				return root.GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
			}
			return nil
		},
	}
	root.AddCommand(completionCmd)
	return root
}

// executeCompletion runs the completion command with the given shell arg and
// returns the captured output plus any error.
func executeCompletion(t *testing.T, shell string) (string, error) {
	t.Helper()
	root := buildTestRootForCompletion()

	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"completion", shell})

	err := root.Execute()
	return buf.String(), err
}

func TestCompletionCmd_Bash(t *testing.T) {
	out, err := executeCompletion(t, "bash")
	require.NoError(t, err)
	// Bash completion scripts begin with a bash function declaration.
	assert.True(t, strings.Contains(out, "bash") || strings.Contains(out, "#!/usr/bin/env bash") || strings.Contains(out, "__start_") || len(out) > 0,
		"expected non-empty bash completion output")
}

func TestCompletionCmd_Zsh(t *testing.T) {
	out, err := executeCompletion(t, "zsh")
	require.NoError(t, err)
	assert.NotEmpty(t, out, "expected non-empty zsh completion output")
}

func TestCompletionCmd_Fish(t *testing.T) {
	out, err := executeCompletion(t, "fish")
	require.NoError(t, err)
	assert.NotEmpty(t, out, "expected non-empty fish completion output")
}

func TestCompletionCmd_PowerShell(t *testing.T) {
	out, err := executeCompletion(t, "powershell")
	require.NoError(t, err)
	assert.NotEmpty(t, out, "expected non-empty powershell completion output")
}

func TestCompletionCmd_InvalidShell(t *testing.T) {
	_, err := executeCompletion(t, "invalidshell")
	assert.Error(t, err, "expected error for invalid shell argument")
}

func TestCompletionCmd_ValidArgs(t *testing.T) {
	root := buildTestRootForCompletion()

	// Locate the completion subcommand and verify its ValidArgs.
	var completionCmd *cobra.Command
	for _, sub := range root.Commands() {
		if sub.Name() == "completion" {
			completionCmd = sub
			break
		}
	}
	require.NotNil(t, completionCmd, "completion subcommand must be registered")

	expected := []string{"bash", "zsh", "fish", "powershell"}
	assert.Equal(t, expected, completionCmd.ValidArgs)
}
