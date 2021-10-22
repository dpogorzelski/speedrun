package cli

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	jsonhandler "github.com/apex/log/handlers/json"
	texthandler "github.com/apex/log/handlers/text"
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

func Execute() {
	var rootCmd = &cobra.Command{
		Use:           "speedrun",
		Short:         "Control your compute fleet at scale",
		Version:       fmt.Sprintf("%s, commit: %s, date: %s", version, commit, date),
		SilenceUsage:  true,
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
	configPath := filepath.Join(dir, "config.toml")

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", configPath, "config file")
	rootCmd.PersistentFlags().StringP("loglevel", "l", "info", "Log level")
	rootCmd.PersistentFlags().BoolP("json", "j", false, "Output logs in JSON format")
	viper.BindPFlag("logging.loglevel", rootCmd.PersistentFlags().Lookup("loglevel"))
	viper.BindPFlag("logging.json", rootCmd.PersistentFlags().Lookup("json"))

	rootCmd.DisableSuggestions = false

	if err := rootCmd.Execute(); err != nil {
		log.Error(err.Error())
	}
}

func initConfig() {
	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv()

	json := viper.GetBool("logging.json")
	if json {
		handler := jsonhandler.New(os.Stdout)
		log.SetHandler(handler)
	} else {
		handler := texthandler.New(os.Stdout)
		log.SetHandler(handler)
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Warnf("Couldn't read config at \"%s\", starting with default settings", viper.ConfigFileUsed())
	}

	lvl, err := log.ParseLevel(viper.GetString("logging.loglevel"))
	if err != nil {
		log.Fatalf("couldn't parse log level: %s (%s)", err, lvl)
		return
	}
	log.SetLevel(lvl)
}
