package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

// Initialized checks if config directory was created
func Initialized(cmd *cobra.Command, args []string) error {
	home, err := homedir.Dir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(home, ".config", "speedrun")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return fmt.Errorf("Try running 'speedrun init' first")
	}
	return nil
}
