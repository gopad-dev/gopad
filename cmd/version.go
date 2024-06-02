package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewVersionCmd(parent *cobra.Command, version string, commit string) {
	cmd := &cobra.Command{
		Use:     "version",
		Short:   "print gopad version",
		Example: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("gopad: %s (%s)\n", version, commit)
			return nil
		},
	}

	parent.AddCommand(cmd)
}
