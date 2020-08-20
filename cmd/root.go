package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var cfgFile string
var verbosity string

var rootCmd = &cobra.Command{
	Use:   "executor",
	Short: "Execute commands at scale",
	Long: `Executor is an application that allows command execution at scale.

executor run "ls -l"`,
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.executor/config)")
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
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			log.Fatal(err)
		}

		viper.SetConfigName("config")
		viper.AddConfigPath(filepath.Join(home, ".executor"))
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}
	setUpLogs(viper.GetString("verbosity"))
}

func setUpLogs(level string) error {
	lvl, err := log.ParseLevel(level)
	if err != nil {
		return err
	}
	log.SetLevel(lvl)
	formatter := &prefixed.TextFormatter{
		DisableTimestamp: true,
	}
	// formatter.SetColorScheme(&prefixed.ColorScheme{
	// 	PrefixStyle: "cyan",
	// })
	log.SetFormatter(formatter)
	return nil
}
