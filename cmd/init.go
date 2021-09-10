package cmd

import (
	"github.com/speedrunsh/speedrun/config"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize speedrun",
	RunE: func(cmd *cobra.Command, args []string) error {
		return config.Create()
	},
}

func init() {
	initCmd.SetUsageTemplate(usage)
}
