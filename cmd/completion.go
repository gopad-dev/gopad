package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var completionInstallCommands = map[string]string{
	"bash": `This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

### Linux
$ ${name} completion bash -i=/etc/bash_completion.d/${name}

### macOS
$ ${name} completion bash -i=$(brew --prefix)/etc/bash_completion.d/${name}`,
	"zsh": `Enable bash completion in Zsh:
% echo "autoload -U compinit; compinit" >> ~/.zshrc

### Linux:
% ${name} completion zsh -i="${fpath[1]}/_${name}"
### macOS:
% ${name} completion zsh -i=$(brew --prefix)/share/zsh/site-functions/_${name}`,
	"fish": `Run the following command to enable fish completion:
> ${name} completion fish -i=~/.config/fish/completions/${name}.fish`,
	"powershell": `Run the following command to enable powershell completion:
> ${name} completion powershell | Out-String | Invoke-Expression`,
}

func NewCompletionCmd(parent *cobra.Command) {
	parent.InitDefaultCompletionCmd()
	var cmd *cobra.Command
	for _, child := range parent.Commands() {
		if child.Name() == "completion" {
			cmd = child
			break
		}
	}
	if cmd == nil {
		return
	}
	noDesc := parent.CompletionOptions.DisableDescriptions

	for _, child := range cmd.Commands() {
		child.RunE = func(cmd *cobra.Command, args []string) error {
			shellName := cmd.Name()
			shouldPrint, _ := cmd.Flags().GetBool("print")
			if shouldPrint {
				switch shellName {
				case "bash":
					return cmd.Root().GenBashCompletionV2(parent.OutOrStdout(), !noDesc)
				case "zsh":
					return cmd.Root().GenZshCompletion(parent.OutOrStdout())
				case "fish":
					return cmd.Root().GenFishCompletion(parent.OutOrStdout(), !noDesc)
				case "powershell":
					return cmd.Root().GenPowerShellCompletion(parent.OutOrStdout())
				}
				return fmt.Errorf("unsupported shell: %s", shellName)
			}

			filename, _ := cmd.Flags().GetString("install")
			if filename != "" {
				switch shellName {
				case "bash":
					return cmd.Root().GenBashCompletionFileV2(filename, !noDesc)
				case "zsh":
					return cmd.Root().GenZshCompletionFile(filename)
				case "fish":
					return cmd.Root().GenFishCompletionFile(filename, !noDesc)
				case "powershell":
					return cmd.Root().GenPowerShellCompletionFile(filename)
				}
				return fmt.Errorf("unsupported shell: %s", shellName)
			}

			cmd.Println(strings.ReplaceAll(completionInstallCommands[shellName], "${name}", cmd.Root().Name()))
			return nil
		}
	}

	cmd.PersistentFlags().BoolP("print", "p", false, "prints the completion script to stdout")
	cmd.PersistentFlags().StringP("install", "i", "", "installs the completion script to the specified location")
}
