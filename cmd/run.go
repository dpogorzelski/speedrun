package cmd

import (
	"fmt"
	"os"
	"sync"

	"speedrun/gcp"
	"speedrun/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run commands on GCE instances",
	Args:  cobra.ExactArgs(1),
	RunE:  run,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := utils.ConfigInitialized()
		if err != nil {
			return err
		}
		err = gcp.ComputeInit()
		return err
	},
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

	p := utils.NewProgress()
	p.Start("Fetching list of GCE instances")
	instances, err := gcp.GetInstances(project, filter)
	if err != nil {
		p.Error(err)
	}
	if len(instances) == 0 {
		p.Failure("no instances found")
	}
	p.Stop()

	p.Start("Updating project metadata")
	err = gcp.UpdateProjectMetadata(project, pubKey)
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
			go gcp.UpdateInstanceMetadata(&wg, project, instances[a+i], pubKey)
		}
		wg.Wait()
	}
	p.Stop()

	p.Start(fmt.Sprintf("Running [%s]", args[0]))
	result, err := utils.Execute(args[0], instances, privKey)
	if err != nil {
		p.Error(err)
		os.Exit(1)
	}
	p.Stop()
	result.PrintResult(onlyFailures)
	return nil
}
