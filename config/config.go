package config

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"

	"github.com/apex/log"
	"github.com/manifoldco/promptui"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

//Create generates a new config
func Create() error {
	viper.SetDefault("loglevel", "info")
	project := promptui.Prompt{
		Label: "Project id",
		Validate: func(input string) error {
			match, err := regexp.MatchString("^[a-z][a-z0-9-]{6,30}", input)
			if err != nil {
				return err
			}
			if !match {
				return errors.New("invalid projectid")
			}
			return nil
		},
	}

	prompt := promptui.Select{
		Label:    "Pick a cloud provider",
		Items:    []string{"Google Cloud"},
		HideHelp: true,
	}

	_, result, err := prompt.Run()
	if err != nil {
		log.Fatal(err.Error())
	}

	switch result {
	case "Google Cloud":
		result, err := project.Run()
		if err != nil {
			log.Fatal(err.Error())
		}
		viper.SetDefault("gcp.projectid", result)
	}

	if err != nil {
		log.Fatal(err.Error())
	}

	home, err := homedir.Dir()
	if err != nil {
		return err
	}

	dir := filepath.Join(home, ".speedrun")
	path := filepath.Join(dir, "config.toml")

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, 0755)
		if err != nil {
			return err
		}
	}

	err = viper.WriteConfigAs(path)
	if err != nil {
		return err
	}
	return nil
}
