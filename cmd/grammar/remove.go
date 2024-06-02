package grammar

import (
	"log"

	"github.com/spf13/cobra"
)

func NewRemoveCmd(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:     "remove",
		Short:   "Used to remove your tree sitter grammars",
		Example: "gopad grammar remove go",
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			configDir, _ := cmd.Flags().GetString("config-dir")

			log.Println(configDir)
			return nil
		},
	}

	parent.AddCommand(cmd)
}
