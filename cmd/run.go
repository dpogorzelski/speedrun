package cmd

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"

	"nyx/gcp"
	"nyx/helpers"

	"github.com/briandowns/spinner"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run commands on GCE instances",
	Args:  cobra.ExactArgs(1),
	Run:   run,
	PreRun: func(cmd *cobra.Command, args []string) {
		home, err := homedir.Dir()
		if err != nil {
			log.Fatal(err)
		}

		configDir := filepath.Join(home, ".nyx")
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			log.Fatal("Try running 'nyx init' first")
		}
	},
}

func init() {
	var filter string

	rootCmd.AddCommand(runCmd)
	runCmd.PersistentFlags().StringVar(&filter, "filter", "", "gcloud resource filter")
}

func run(cmd *cobra.Command, args []string) {
	project := viper.GetString("project")

	filter, err := cmd.Flags().GetString("filter")
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	forceNewKey, err := cmd.Flags().GetBool("force-new-key")
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	pubKey, privKey, err := helpers.GetKeyPair(forceNewKey)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	green := color.New(color.FgGreen).SprintfFunc()
	red := color.New(color.FgRed).SprintfFunc()
	yellow := color.New(color.FgYellow).SprintfFunc()
	command := color.New(color.FgCyan).SprintFunc()
	tag := "·"

	s.Suffix = " Fetching list of GCE instances"
	s.FinalMSG = green("%s Fetching list of GCE instances\n", tag)
	s.Start()
	instances, err := gcp.GetInstances(project, filter)
	if err != nil {
		s.FinalMSG = red("%s Fetching list of GCE instances: %s\n", tag, err)
		s.Stop()
		os.Exit(1)
	}
	if len(instances) == 0 {
		s.FinalMSG = yellow("%s Fetching list of GCE instances: no instances found, check and/or relax your --filter settings\n", tag)
		s.Stop()
		os.Exit(0)
	}
	s.Stop()

	s.Suffix = " Updating GCE metadata"
	s.FinalMSG = green("%s Updating GCE metadata\n", tag)
	s.Restart()
	var wg sync.WaitGroup
	for _, instance := range instances {
		wg.Add(1)
		go gcp.UpdateInstanceMetadata(&wg, project, instance, pubKey)
	}
	wg.Wait()
	s.Stop()

	s.Suffix = fmt.Sprintf(" Running [%s]", command(args[0]))
	s.FinalMSG = green("%s Running [%s]:\n", tag, command(args[0]))
	s.Restart()
	result, err := helpers.Execute(args[0], instances, privKey)
	if err != nil {
		s.FinalMSG = red("%s Running [%s]: %s\n", tag, command(args[0]), err)
		s.Stop()
		os.Exit(1)
	}
	s.Stop()
	result.PrintResult()
	//clearProjectMetadata()
}
