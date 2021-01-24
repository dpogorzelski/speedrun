package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"github.com/tcnksm/go-input"
	"github.com/urfave/cli/v2"
)

type config struct {
	Gcp gcpConfig `toml:"gcp"`
}

type gcpConfig struct {
	ProjectID string `toml:"projectid"`
}

func initialize(ctx *cli.Context) error {
	err := createConfig()
	if err != nil {
		return err
	}

	return nil
}

func loadConfig(configDir string) error {
	viper.SetConfigName("config.toml")
	viper.SetConfigType("toml")
	viper.AddConfigPath(configDir)

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	return nil
}

func configDir() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".config", "speedrun"), nil
}

func createConfig() error {
	var err error
	ui := &input.UI{}
	config := &config{Gcp: gcpConfig{}}

	config.Gcp.ProjectID, err = ui.Ask("Google Cloud project ID?", &input.Options{
		Required: true,
		Loop:     true,
	})
	if err != nil {
		return err
	}

	dir, err := configDir()
	if err != nil {
		return err
	}

	file := filepath.Join(dir, "config.toml")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0700)
	}

	f, err := os.Create(file)
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

func configInitialized(c *cli.Context) error {
	configDir, err := configDir()
	if err != nil {
		return cli.Exit(fmt.Errorf("Try running 'speedrun init' first"), 1)
	}

	return loadConfig(configDir)
}
