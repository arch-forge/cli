package cli

import "github.com/spf13/cobra"

// newCompletionCmd creates the completion command that generates shell completion scripts.
func newCompletionCmd(rootCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `To load completions:

Bash:
  $ source <(arch_forge completion bash)
  # To load completions for each session, add this to your ~/.bashrc:
  $ arch_forge completion bash > ~/.bash_completion.d/arch_forge

Zsh:
  $ arch_forge completion zsh > "${fpath[1]}/_arch_forge"
  # or
  $ source <(arch_forge completion zsh)

Fish:
  $ arch_forge completion fish > ~/.config/fish/completions/arch_forge.fish

PowerShell:
  $ arch_forge completion powershell | Out-String | Invoke-Expression`,
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return rootCmd.GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				return rootCmd.GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				return rootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				return rootCmd.GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
			}
			return nil
		},
	}
	return cmd
}
