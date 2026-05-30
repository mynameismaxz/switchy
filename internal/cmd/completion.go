package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/mynameismaxz/switchy/internal/profile"
)

func newCompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion script for swy",
		Long: `Generate shell completion script for swy.

To load completions:

  Bash:
    source <(swy completion bash)
    # To load permanently:
    swy completion bash >> ~/.bashrc

  Zsh:
    source <(swy completion zsh)
    # To load permanently:
    swy completion zsh >> ~/.zshrc

  Fish:
    swy completion fish > ~/.config/fish/completions/swy.fish

  PowerShell:
    swy completion powershell | Out-String | Invoke-Expression
`,
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletionV2(os.Stdout, true)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
			return nil
		},
	}
}

// profileCompletion returns profile names for shell autocomplete.
// Used as ValidArgsFunction on commands that accept a profile name argument.
func profileCompletion(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	names, current, err := profile.ListProfiles()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	comps := make([]cobra.Completion, 0, len(names))
	for _, n := range names {
		if n == current {
			comps = append(comps, cobra.CompletionWithDesc(n, "current"))
		} else {
			comps = append(comps, n)
		}
	}
	return comps, cobra.ShellCompDirectiveNoFileComp
}
