package config

import (
	"os"
	"path/filepath"

	"github.com/apex/log"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"github.com/tcnksm/go-input"
)

//Create generates a new config
func Create() error {
	viper.SetDefault("loglevel", "info")
	log.Info("It seems like you're new to Speedrun, let's get you up and running.")

	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}

	b, err := ui.Select("Pick a cloud provider", []string{"Google Cloud"}, &input.Options{
		Required: true,
		Loop:     true,
	})

	switch b {
	case "Google Cloud":
		project, err := ui.Ask("What's the project id?", &input.Options{
			Required: true,
			Loop:     true,
		})
		if err != nil {
			log.Fatal(err.Error())
		}
		viper.SetDefault("gcp.projectid", project)
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
