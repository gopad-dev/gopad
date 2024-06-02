package grammar

import (
	"log"

	"github.com/spf13/cobra"
)

func NewUpdateCmd(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:     "update",
		Short:   "Used to update your tree sitter grammars",
		Example: "gopad grammar update go",
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			configDir, _ := cmd.Flags().GetString("config-dir")

			log.Println(configDir)
			return nil
		},
	}

	parent.AddCommand(cmd)
}
