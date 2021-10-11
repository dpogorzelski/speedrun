package config

import (
	"os"
	"path/filepath"

	"github.com/apex/log"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

//Create generates a new config
func Create() error {
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
	log.Infof("Your config was saved at: %s", path)
	return nil
}
