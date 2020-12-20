package cmd

import (
	"fmt"
	"os"
	"sync"

	"speedrun/gcp"
	"speedrun/utils"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var runCmd = &cobra.Command{
	Use:                   "run",
	Short:                 "Run commands on GCE instances",
	Args:                  cobra.ExactArgs(1),
	RunE:                  run,
	PreRunE:               utils.ConfigInitialized,
	DisableFlagsInUseLine: true,
}

func init() {
	var filter string
	var onlyFailures bool

	rootCmd.AddCommand(runCmd)
	runCmd.PersistentFlags().StringVar(&filter, "filter", "", "gcloud resource filter")
	runCmd.PersistentFlags().BoolVar(&onlyFailures, "only-failures", false, "print only failures and errors")
}

func run(cmd *cobra.Command, args []string) error {
	project := viper.GetString("gcp.projectid")

	client, err := gcp.NewComputeClient(project)
	if err != nil {
		return err
	}

	filter, err := cmd.Flags().GetString("filter")
	if err != nil {
		return err
	}

	onlyFailures, err := cmd.Flags().GetBool("only-failures")
	if err != nil {
		return err
	}

	pubKey, privKey, err := utils.GetKeyPair()
	if err != nil {
		return err
	}

	if err = client.GetFirewallRules(); err != nil {
		return err
	}

	p := utils.NewProgress()
	p.Start("Fetching list of GCE instances")
	instances, err := client.GetInstances(filter)
	if err != nil {
		p.Error(err)
	}
	if len(instances) == 0 {
		p.Failure("no instances found")
	}
	p.Stop()

	p.Start("Updating project metadata")
	err = client.UpdateProjectMetadata(pubKey)
	if err != nil {
		p.Error(err)
	}
	p.Stop()

	p.Start("Updating instance metadata")
	batch := 50
	for i := 0; i < len(instances); i += batch {
		j := i + batch
		if j > len(instances) {
			j = len(instances)
		}
		var wg sync.WaitGroup
		for a := range instances[i:j] {
			wg.Add(1)
			go client.UpdateInstanceMetadata(&wg, instances[a+i], pubKey)
		}
		wg.Wait()
	}
	p.Stop()

	p.Start(fmt.Sprintf("Running [%s]", color.BlueString(args[0])))
	result, err := utils.Execute(args[0], instances, privKey)
	if err != nil {
		p.Error(err)
		os.Exit(1)
	}
	p.Stop()
	result.PrintResult(onlyFailures)
	return nil
}
