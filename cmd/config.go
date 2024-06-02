package cmd

import (
	"embed"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"go.gopad.dev/gopad/gopad/config"
)

func NewConfigCmd(parent *cobra.Command, defaultConfigs embed.FS) {
	cmd := &cobra.Command{
		Use:     "config",
		Short:   "Used to create a default config in the target location",
		Example: "gopad config ~/.config/gopad",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := args[0]

			if err := config.Create(configPath, defaultConfigs); err != nil {
				log.Panicln("failed to create config:", err)
			}
			fmt.Println("created config in", configPath)
			return nil
		},
	}

	parent.AddCommand(cmd)
}

func initConfig(configDir string, defaultConfigs embed.FS) {
	if configDir == "" {
		var err error
		configDir, err = config.FindHome()
		if err != nil {
			log.Panicln("failed to find config dir:", err)
		}
	}

	if err := config.Load(configDir, defaultConfigs); err != nil {
		log.Panicln("failed to load config:", err)
	}
}
