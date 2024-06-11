package cmd

import (
	"github.com/spf13/cobra"

	"go.gopad.dev/gopad/cmd/grammar"
)

func NewGrammarCmd(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:     "grammar",
		Short:   "Used to manage your tree sitter grammar installations",
		Long:    "",
		Example: "",
	}

	parent.AddCommand(cmd)

	grammar.NewInstallCmd(cmd)
	grammar.NewListCmd(cmd)
	grammar.NewRemoveCmd(cmd)
	grammar.NewUpdateCmd(cmd)
}
