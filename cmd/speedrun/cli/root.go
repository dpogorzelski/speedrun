package cli

import (
	_ "embed"
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

//go:embed templates/root.tmpl
var rootUsage string

//go:embed templates/usage.tmpl
var usage string

//Execute runs the root command
func Execute() {
	var rootCmd = &cobra.Command{
		Use:           "speedrun",
		Short:         "Control your compute fleet at scale",
		Version:       fmt.Sprintf("%s, commit: %s, date: %s", version, commit, date),
		SilenceUsage:  false,
		SilenceErrors: true,
	}

	cobra.OnInitialize(initConfig)
	rootCmd.SetUsageTemplate(rootUsage)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(serviceCmd)

	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err.Error())
	}
	dir := filepath.Join(home, ".speedrun")
	path := filepath.Join(dir, "config.toml")

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", path, "config file")
	rootCmd.PersistentFlags().StringP("loglevel", "l", "info", "Log level")
	viper.BindPFlag("loglevel", rootCmd.PersistentFlags().Lookup("loglevel"))

	rootCmd.DisableSuggestions = false

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}

func initConfig() {
	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Warnf("Config file not found at \"%s\", starting with default settings", viper.ConfigFileUsed())
		} else {
			log.Warnf("Couldn't read config at \"%s\", starting with default settings", viper.ConfigFileUsed())
		}
	}

	lvl, err := log.ParseLevel(viper.GetString("loglevel"))
	if err != nil {
		log.Fatalf("couldn't parse log level: %s (%s)", err, lvl)
		return
	}
	log.SetLevel(lvl)
}
