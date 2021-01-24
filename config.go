package main

import (
	"bytes"
	"os"

	"github.com/pelletier/go-toml"
	"github.com/spf13/viper"
	"github.com/tcnksm/go-input"
	"github.com/urfave/cli/v2"
)

// Config type holds the configuration struct
type Config struct {
	Gcp gcpConfig
}

type gcpConfig struct {
	ProjectID string
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

	c.Gcp.ProjectID, err = ui.Ask("Google Cloud project ID?", &input.Options{
		Required: true,
		Loop:     true,
	})
	if err != nil {
		return err
	}

	b, err := toml.Marshal(c)
	viper.ReadConfig(bytes.NewBuffer(b))

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, mode)
	}

	err = viper.WriteConfigAs("/Users/dpogorzelski/.speedrun/config.toml")
	if err != nil {
		return err
	}
	return nil
}
