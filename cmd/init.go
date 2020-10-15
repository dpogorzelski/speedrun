package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"log"

	"github.com/AlecAivazis/survey/v2"
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
	projectID := ""
	prompt := &survey.Input{Message: "Google Cloud project ID?"}

	// validate := func(val interface{}) error {
	// 	if str, ok := val.(string); !ok || len(str) > 30 || len(str) > 6 {
	// 		return errors.New("This response cannot be longer than 10 characters.")
	// 	}
	// 	return nil
	// }

	err := survey.AskOne(prompt, &projectID, survey.WithValidator(survey.Required))
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	config := make(map[string]string)

	config["project"] = projectID
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
