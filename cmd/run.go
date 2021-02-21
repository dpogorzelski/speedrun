package cmd

import (
	"fmt"
	gcp "speedrun/cloud"
	"speedrun/marathon"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var runCmd = &cobra.Command{
	Use:     "run <command to run>",
	Short:   "Run command on remote servers",
	Example: "  speedrun run whoami -r\n  speedrun run whoami --only-failures --filter \"labels.foo = bar AND labels.environment = staging\"",
	Args:    cobra.MinimumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
	RunE: run,
}

func init() {
	runCmd.Flags().String("filter", "", "Fetch instances that match the filter")
	runCmd.Flags().String("projectid", "", "Override GCP project id")
	runCmd.Flags().Bool("only-failures", false, "Print only failures and errors")
	runCmd.Flags().Bool("ignore-fingerprint", false, "Ignore host's fingerprint mismatch")
	// runCmd.Flags().Bool("use-private-ip", false, "Connect to private IPs instead of public ones")
}

func run(cmd *cobra.Command, args []string) error {
	command := strings.Join(args, " ")
	filter, err := cmd.Flags().GetString("filter")
	if err != nil {
		return err
	}

	onlyFailures, err := cmd.Flags().GetBool("only-failures")
	if err != nil {
		return err
	}

	ignoreFingerprint, err := cmd.Flags().GetBool("ignore-fingerprint")
	if err != nil {
		return err
	}

	client, err := gcp.NewComputeClient(viper.GetString("gcp.projectid"))
	if err != nil {
		return err
	}

	privateKeyPath, err := determineKeyFilePath()
	if err != nil {
		return err
	}

	log.Info("Fetching list of GCE instances")
	instances, err := client.GetInstances(filter)
	if err != nil {
		return err
	}
	if len(instances) == 0 {
		log.Warn("No instances found")
		return nil
	}

	log.Info(fmt.Sprintf("Running [%s]", color.BlueString(command)))
	timeout, err := time.ParseDuration("10s")
	if err != nil {
		return err
	}

	m := marathon.New(command, timeout)
	instanceDict := map[string]string{}
	for _, instance := range instances {
		instanceDict[instance.NetworkInterfaces[0].AccessConfigs[0].NatIP] = instance.Name
	}
	err = m.Run(instanceDict, privateKeyPath, ignoreFingerprint)
	if err != nil {
		return err
	}

	m.PrintResult(onlyFailures)
	return nil
}
