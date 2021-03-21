package cmd

import (
	"speedrun/config"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize speedrun",
	RunE: func(cmd *cobra.Command, args []string) error {
		return config.Create()
	},
}
