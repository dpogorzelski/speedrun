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
		return cli.Exit(err, 1)
	}

	err = createKey(ctx)
	if err != nil {
		return cli.Exit(err, 1)
	}
	return nil
}

func loadConfig(ctx *cli.Context) error {
	home, err := homedir.Dir()
	if err != nil {
		cli.Exit(err, 1)
	}

	configDir := filepath.Join(home, ".config", "speedrun")
	viper.SetConfigName("config.toml")
	viper.SetConfigType("toml")
	viper.AddConfigPath(configDir)

	err = viper.ReadInConfig()
	if err != nil {
		cli.Exit(err, 1)
	}

	return nil
	// setUpLogs(viper.GetString("verbosity"))
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

	home, err := homedir.Dir()
	if err != nil {
		return err
	}

	path := filepath.Join(home, ".config", "speedrun", "config.toml")
	if _, err := os.Stat(filepath.Join(home, ".config", "speedrun")); os.IsNotExist(err) {
		os.Mkdir(filepath.Join(home, ".config", "speedrun"), 0700)
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

func configInitialized(ctx *cli.Context) error {
	home, err := homedir.Dir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(home, ".config", "speedrun")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return cli.Exit(fmt.Errorf("Try running 'speedrun init' first"), 1)
	}
	return nil
}
