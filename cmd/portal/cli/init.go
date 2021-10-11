package cli

import (
	"github.com/speedrunsh/speedrun/pkg/common/config"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize portal",
	RunE: func(cmd *cobra.Command, args []string) error {
		return config.Create()
	},
}
