package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"

	"speedrun/helpers"

	"github.com/spf13/cobra"
)

var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Manage ssh keys",
	Args:  cobra.ExactArgs(1),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		home, err := homedir.Dir()
		if err != nil {
			helpers.Error(err.Error())
		}

		configDir := filepath.Join(home, ".config", "speedrun")
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			helpers.Error("Try running 'speedrun init' first")
		}
	},
}

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Generates a new ssh key",
	Run:   new,
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Shows private key",
	Run:   show,
}

func init() {
	rootCmd.AddCommand(keyCmd)
	keyCmd.AddCommand(newCmd)
	keyCmd.AddCommand(showCmd)
}

func new(cmd *cobra.Command, args []string) {
	fmt.Println("generated new ssh key")
	helpers.GenerateKeyPair()
}

func show(cmd *cobra.Command, args []string) {
	fmt.Println("showing private key")
}
