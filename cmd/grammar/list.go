package grammar

import (
	"log"

	"github.com/spf13/cobra"
)

func NewListCmd(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "Used to list your tree sitter grammars",
		Example: "gopad grammar list go",
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			configDir, _ := cmd.Flags().GetString("config-dir")

			log.Println(configDir)
			return nil
		},
	}

	parent.AddCommand(cmd)
}
