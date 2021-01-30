package main

import (
	"bytes"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pelletier/go-toml"
	"github.com/spf13/viper"
	"github.com/tcnksm/go-input"
	"github.com/urfave/cli/v2"
)

// Config type holds the configuration struct
type Config struct {
	Gcp struct {
		Projectid string
	}
}

// NewConfig initializes viper
func NewConfig() (*Config, error) {
	c := &Config{}

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath("$HOME/.speedrun")

	return c, nil
}

// Read will read the config file if exists and unmarshal it to the Config type
func (c *Config) Read(ctx *cli.Context) error {
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	err = viper.Unmarshal(&c)
	if err != nil {
		return err
	}

	return nil
}

// Create will create a new config file if it doesn't exist
func (c *Config) Create(ctx *cli.Context) error {
	var err error
	ui := &input.UI{}

	c.Gcp.Projectid, err = ui.Ask("Google Cloud project ID?", &input.Options{
		Required: true,
		Loop:     true,
	})
	if err != nil {
		return err
	}

	b, err := toml.Marshal(c)
	viper.ReadConfig(bytes.NewBuffer(b))

	home, err := homedir.Dir()
	if err != nil {
		return err
	}

	dir := filepath.Join(home, ".speedrun")
	path := filepath.Join(dir, "config.toml")

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, os.ModeDir)
	}

	err = viper.WriteConfigAs(path)
	if err != nil {
		return err
	}
	return nil
}
