package cmd

import (
	"embed"

	"github.com/spf13/cobra"

	"go.gopad.dev/gopad/cmd/grammar"
)

func NewGrammarCmd(parent *cobra.Command, defaultConfigs embed.FS) {
	cmd := &cobra.Command{
		Use:     "grammar",
		Short:   "Manage Tree-Sitter grammars",
		Long:    "",
		Example: "",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			configDir, _ := cmd.Flags().GetString("config-dir")
			loadConfig(configDir, defaultConfigs)
		},
	}

	parent.AddCommand(cmd)

	grammar.NewInstallCmd(cmd)
	grammar.NewListCmd(cmd)
	grammar.NewRemoveCmd(cmd)
	grammar.NewUpdateCmd(cmd)
}
