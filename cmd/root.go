package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/apex/log"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var version string
var commit string
var date string

//Execute runs the root command
func Execute() {
	// cobra.OnInitialize(initConfig)
	var rootCmd = &cobra.Command{
		Use:           "speedrun",
		Short:         "Cloud first command execution",
		Version:       fmt.Sprintf("%s, commit: %s, date: %s", version, commit, date),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(keyCmd)
	rootCmd.AddCommand(runCmd)

	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err.Error())
	}
	dir := filepath.Join(home, ".speedrun")
	path := filepath.Join(dir, "config.toml")

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", path, "config file")
	rootCmd.PersistentFlags().StringP("loglevel", "l", "info", "Log level")
	viper.BindPFlag("loglevel", rootCmd.PersistentFlags().Lookup("loglevel"))

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}

func initConfig() {
	dir, file := filepath.Split(cfgFile)
	viper.SetConfigName(file)
	viper.SetConfigType("toml")
	viper.AddConfigPath(dir)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err != nil {
				log.Error(err.Error())
				log.Fatal("Run `speedrun init` first")
			}
		} else {
			log.Fatal(err.Error())
		}
	}
	log.SetLevelFromString(viper.GetString("loglevel"))
}
