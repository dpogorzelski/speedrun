package cmd

import (
	"path"

	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

//Execute runs the root command
func Execute() {
	// cobra.OnInitialize(initConfig)
	var rootCmd = &cobra.Command{
		Use:     "speedrun",
		Short:   "Cloud first command execution",
		Version: "0.1.0",
	}

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(keyCmd)
	rootCmd.AddCommand(runCmd)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "$HOME/.speedrun/config.toml", "config file")
	rootCmd.PersistentFlags().StringP("loglevel", "l", "info", "Log level")
	viper.BindPFlag("loglevel", rootCmd.PersistentFlags().Lookup("loglevel"))

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}

func initConfig() {
	dir, file := path.Split(cfgFile)
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
