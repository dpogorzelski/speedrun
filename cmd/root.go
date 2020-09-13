package cmd

import (
	"fmt"
	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"os"
	"path/filepath"
)

var cfgFile string
var verbosity string

var rootCmd = &cobra.Command{
	Use:   "nyx",
	Short: "Execute commands at scale",
	Long:  `Nyx is an application that allows command execution at scale.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	var project string
	var keyPath string
	var forceNewKey bool

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&verbosity, "verbosity", "v", log.InfoLevel.String(), "Log level (debug, info, warn, error, fatal, panic)")
	viper.BindPFlag("verbosity", rootCmd.PersistentFlags().Lookup("verbosity"))

	rootCmd.PersistentFlags().StringVar(&keyPath, "key-path", "", "path to the private SSH key to use")
	viper.BindPFlag("key-path", rootCmd.PersistentFlags().Lookup("key-path"))

	rootCmd.PersistentFlags().BoolVar(&forceNewKey, "force-new-key", false, "force creation of a new SSH key pair")
	viper.BindPFlag("force-new-key", rootCmd.PersistentFlags().Lookup("force-new-key"))

	rootCmd.PersistentFlags().StringVar(&project, "project", "", "google cloud project id")
	viper.BindPFlag("project", rootCmd.PersistentFlags().Lookup("project"))
}

func initConfig() {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}

	configDir := filepath.Join(home, ".nyx")
	viper.SetConfigName("config.toml")
	viper.SetConfigType("toml")
	viper.AddConfigPath(configDir)

	if err := viper.ReadInConfig(); err == nil {
		log.Debug("Using config file ", viper.ConfigFileUsed())
	}
	setUpLogs(viper.GetString("verbosity"))
}

func setUpLogs(level string) error {
	lvl, err := log.ParseLevel(level)
	if err != nil {
		return err
	}
	log.SetLevel(lvl)
	// log.SetOutput()
	formatter := &prefixed.TextFormatter{
		DisableTimestamp: true,
		FullTimestamp:    true,
	}
	// formatter.SetColorScheme(&prefixed.ColorScheme{
	// 	PrefixStyle: "cyan",
	// })
	log.SetFormatter(formatter)
	return nil
}
