package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"log"

	"github.com/BurntSushi/toml"
	homedir "github.com/mitchellh/go-homedir"

	"nyx/helpers"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize nyx",
	Run:   initialize,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func initialize(cmd *cobra.Command, args []string) {
	err := createConfig()
	if err != nil {
		log.Fatal(err.Error())
	}

	_, _, err = helpers.GenerateKeyPair()
	if err != nil {
		log.Fatal(err.Error())
	}
}

func createConfig() error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Google Cloud project id: ")
	config := make(map[string]string)

	projectID, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	config["project"] = strings.TrimSpace(projectID)
	home, err := homedir.Dir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".config", "nyx", "config.toml")
	if _, err := os.Stat(filepath.Join(home, ".config", "nyx")); os.IsNotExist(err) {
		os.Mkdir(filepath.Join(home, ".config", "nyx"), 0700)
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := toml.NewEncoder(f).Encode(config); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	return nil
}