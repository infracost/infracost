package main

import (
	"os"

	"github.com/spf13/cobra"
)

func completionCmd() *cobra.Command {
	completionCmd := &cobra.Command{
		Use:   "completion [bash | zsh | fish | powershell]",
		Short: "Generate completion script",
		Long: `To load completions:
	
	Bash:
	
		$ source <(infracost completion bash)
	
		# To load completions for each session, execute once:
		# Linux:
		$ infracost completion bash > /etc/bash_completion.d/infracost
		# macOS:
		$ infracost completion bash > /usr/local/etc/bash_completion.d/infracost
	
	Zsh:
	
		# If shell completion is not already enabled in your environment,
		# you will need to enable it.  You can execute the following once:
	
		$ echo "autoload -U compinit; compinit" >> ~/.zshrc
	
		# To load completions for each session, execute once:
		$ infracost completion zsh > "${fpath[1]}/_infracost"
	
		# You will need to start a new shell for this setup to take effect.
	
	fish:
	
		$ infracost completion fish | source
	
		# To load completions for each session, execute once:
		$ infracost completion fish > ~/.config/fish/completions/infracost.fish
	
	PowerShell:
	
		PS> infracost completion powershell | Out-String | Invoke-Expression
	
		# To load completions for every new session, run:
		PS> infracost completion powershell > infracost.ps1
		# and source this file from your PowerShell profile.
	`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				_ = cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				_ = cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				_ = cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				_ = cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	}

	return completionCmd
}
