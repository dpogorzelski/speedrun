package cmd

import (
	"os"

	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

//Execute runs the root command
func Execute() {
	// cobra.OnInitialize(initConfig)
	var rootCmd = &cobra.Command{
		Use:   "speedrun",
		Short: "Cloud first command execution",
	}

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(keyCmd)
	rootCmd.AddCommand(runCmd)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.speedrun/config.yaml)")
	rootCmd.PersistentFlags().StringP("loglevel", "l", "info", "Log level")
	viper.BindPFlag("loglevel", rootCmd.PersistentFlags().Lookup("loglevel"))

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("toml")
		viper.AddConfigPath("$HOME/.speedrun")
	}

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
