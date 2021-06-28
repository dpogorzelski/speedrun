package cmd

import (
	"fmt"
	"speedrun/cloud"
	"speedrun/colors"
	"speedrun/key"
	"speedrun/marathon"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var runCmd = &cobra.Command{
	Use:     "run <command to run>",
	Short:   "Run command on remote servers",
	Example: "  speedrun run whoami\n  speedrun run whoami --only-failures --filter \"labels.foo = bar AND labels.environment = staging\"",
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
	runCmd.Flags().Duration("timeout", time.Duration(10*time.Second), "SSH connection timeout")
	runCmd.Flags().Int("concurrency", 100, "Number of maximum concurrent SSH workers")
	runCmd.Flags().Bool("use-private-ip", false, "Connect to private IPs instead of public ones")
	viper.BindPFlag("gcp.projectid", runCmd.Flags().Lookup("projectid"))
	viper.BindPFlag("ssh.timeout", runCmd.Flags().Lookup("timeout"))
	viper.BindPFlag("ssh.ignore-fingerprint", runCmd.Flags().Lookup("ignore-fingerprint"))
	viper.BindPFlag("ssh.only-failures", runCmd.Flags().Lookup("only-failures"))
	viper.BindPFlag("ssh.concurrency", runCmd.Flags().Lookup("concurrency"))
	viper.BindPFlag("ssh.use-private-ip", runCmd.Flags().Lookup("use-private-ip"))

}

func run(cmd *cobra.Command, args []string) error {
	command := strings.Join(args, " ")
	projectid := viper.GetString("gcp.projectid")
	timeout := viper.GetDuration("ssh.timeout")
	ignoreFingerprint := viper.GetBool("ssh.ignore-fingerprint")
	onlyFailures := viper.GetBool("ssh.only-failures")
	concurrency := viper.GetInt("ssh.concurrency")
	usePrivateIP := viper.GetBool("ssh.use-private-ip")
	filter, err := cmd.Flags().GetString("filter")
	if err != nil {
		return err
	}

	client, err := cloud.NewClient(cloud.SetProject(projectid))
	if err != nil {
		log.Fatal(err.Error())
	}

	path, err := determineKeyFilePath()
	if err != nil {
		log.Fatal(err.Error())
	}

	k, err := key.Read(path)
	if err != nil {
		return err
	}

	log.Info("Fetching list of GCE instances")
	instances, err := client.GetInstances(filter)
	if err != nil {
		log.Fatal(err.Error())
	}
	if len(instances) == 0 {
		log.Warn("No instances found")
		return nil
	}

	log.Info(fmt.Sprintf("Running [%s]", colors.Blue(command)))
	m := marathon.New(command, timeout, concurrency)

	err = m.Run(instances, k, ignoreFingerprint, usePrivateIP)
	if err != nil {
		log.Fatal(err.Error())
	}

	m.PrintResult(onlyFailures)
	return nil
}
