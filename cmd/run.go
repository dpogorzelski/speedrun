package cmd

import (
	"os"
	"sync"

	log "github.com/sirupsen/logrus"

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

	instances, err := gcp.GetInstances(project, filter)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	if len(instances) == 0 {
		log.Info("No instances found")
		os.Exit(0)
	}

	var wg sync.WaitGroup
	for _, instance := range instances {
		wg.Add(1)
		go gcp.UpdateInstanceMetadata(&wg, project, instance, pubKey)
	}
	wg.Wait()

	err = helpers.Execute(args[0], instances, privKey)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	//clearProjectMetadata()
}
