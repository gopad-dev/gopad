package cmd

import (
	"github.com/spf13/cobra"
)

func NewVersionCmd(parent *cobra.Command, version string, commit string) {
	cmd := &cobra.Command{
		Use:     "version [flags]",
		Short:   "Show version information",
		Example: "",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("gopad: %s (%s)\n", version, commit)
		},
	}

	parent.AddCommand(cmd)
}
