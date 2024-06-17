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
		Use:               "config [flags]... [dir]",
		Short:             "Create a new config directory with default config files",
		Long:              "Create a new config directory with default config files in the specified directory\n(Default: $XDG_CONFIG_HOME/gopad or $HOME/.config/gopad)",
		Example:           "gopad config ~/.config/gopad",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: cobra.FixedCompletions(nil, cobra.ShellCompDirectiveFilterDirs),
		RunE: func(cmd *cobra.Command, args []string) error {
			var configHome string
			if len(args) > 0 {
				configHome = args[0]
			} else {
				var err error
				configHome, err = config.FindHome()
				if err != nil {
					return fmt.Errorf("failed to find config home dir: %w", err)
				}
			}

			if err := config.Create(configHome, defaultConfigs); err != nil {
				log.Panicln("failed to create config:", err)
			}
			cmd.Println("created config in", configHome)
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
