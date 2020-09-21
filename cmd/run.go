package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/mitchellh/go-homedir"

	"nyx/gcp"
	"nyx/helpers"

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
			helpers.Error(err.Error())
		}

		configDir := filepath.Join(home, ".config", "nyx")
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			helpers.Error("Try running 'nyx init' first")
		}
	},
}

func init() {
	var filter string
	var onlyFailures bool

	rootCmd.AddCommand(runCmd)
	runCmd.PersistentFlags().StringVar(&filter, "filter", "", "gcloud resource filter")
	runCmd.PersistentFlags().BoolVar(&onlyFailures, "only-failures", false, "print only failures and errors")
}

func run(cmd *cobra.Command, args []string) {
	project := viper.GetString("project")

	filter, err := cmd.Flags().GetString("filter")
	if err != nil {
		helpers.Error(err.Error())
	}

	onlyFailures, err := cmd.Flags().GetBool("only-failures")
	if err != nil {
		helpers.Error(err.Error())
	}

	forceNewKey, err := cmd.Flags().GetBool("force-new-key")
	if err != nil {
		helpers.Error(err.Error())
	}

	pubKey, privKey, err := helpers.GetKeyPair(forceNewKey)
	if err != nil {
		helpers.Error(err.Error())
	}

	p := helpers.NewProgress()
	p.Start("Fetching list of GCE instances")
	instances, err := gcp.GetInstances(project, filter)
	if err != nil {
		p.Error(err)
	}
	if len(instances) == 0 {
		p.Failure("no instances found, check and/or relax your --filter settings")
	}
	p.Stop()

	p.Start("Updating GCE metadata")
	batch := 50
	for i := 0; i < len(instances); i += batch {
		j := i + batch
		if j > len(instances) {
			j = len(instances)
		}
		var wg sync.WaitGroup
		for a := range instances[i:j] {
			wg.Add(1)
			go gcp.UpdateInstanceMetadata(&wg, project, instances[a+i], pubKey)
		}
		wg.Wait()
	}

	p.Stop()

	p.Start(fmt.Sprintf("Running [%s]", args[0]))
	result, err := helpers.Execute(args[0], instances, privKey)
	if err != nil {
		p.Error(err)
		os.Exit(1)
	}
	p.Stop()
	result.PrintResult(onlyFailures)
}
