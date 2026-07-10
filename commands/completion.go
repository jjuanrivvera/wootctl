package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	metaRegistrars = append(metaRegistrars, func(_ *deps) *cobra.Command {
		// Cobra generates a `completion` command automatically, but making it explicit lets
		// us document it and keeps dod-check/spec tooling honest about the surface.
		return &cobra.Command{
			Use:                   "completion [bash|zsh|fish|powershell]",
			Short:                 "Generate shell completion scripts",
			DisableFlagsInUseLine: true,
			ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
			Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
			Example: `  cwctl completion zsh > "${fpath[1]}/_cwctl"
  cwctl completion bash > /etc/bash_completion.d/cwctl
  cwctl completion fish > ~/.config/fish/completions/cwctl.fish`,
			RunE: func(cmd *cobra.Command, args []string) error {
				root := cmd.Root()
				switch args[0] {
				case "bash":
					return root.GenBashCompletionV2(cmd.OutOrStdout(), true)
				case "zsh":
					return root.GenZshCompletion(cmd.OutOrStdout())
				case "fish":
					return root.GenFishCompletion(cmd.OutOrStdout(), true)
				default:
					return root.GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
				}
			},
		}
	})
}
