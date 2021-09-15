package cli

import (
	"context"
	"strings"
	"time"

	"github.com/speedrunsh/speedrun/pkg/common/key"
	"github.com/speedrunsh/speedrun/pkg/common/ssh"
	"github.com/speedrunsh/speedrun/pkg/speedrun/cloud"

	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var runCmd = &cobra.Command{
	Use:     "run <command to run>",
	Short:   "Run a shell command on remote servers",
	Example: "  speedrun run whoami\n  speedrun run whoami --only-failures --target \"labels.foo = bar AND labels.environment = staging\"",
	Args:    cobra.MinimumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
	RunE: run,
}

func init() {
	runCmd.Flags().StringP("target", "t", "", "Fetch instances that match the target selection criteria")
	runCmd.Flags().String("projectid", "", "Override GCP project id")
	runCmd.Flags().Bool("only-failures", false, "Print only failures and errors")
	runCmd.Flags().Bool("ignore-fingerprint", false, "Ignore host's fingerprint mismatch")
	runCmd.Flags().Duration("timeout", time.Duration(10*time.Second), "SSH connection timeout")
	runCmd.Flags().Int("concurrency", 100, "Number of maximum concurrent SSH workers")
	runCmd.Flags().Bool("use-private-ip", false, "Connect to private IPs instead of public ones")
	runCmd.Flags().Bool("use-oslogin", false, "Authenticate via OS Login")
	viper.BindPFlag("gcp.projectid", runCmd.Flags().Lookup("projectid"))
	viper.BindPFlag("gcp.use-oslogin", runCmd.Flags().Lookup("use-oslogin"))
	viper.BindPFlag("ssh.timeout", runCmd.Flags().Lookup("timeout"))
	viper.BindPFlag("ssh.ignore-fingerprint", runCmd.Flags().Lookup("ignore-fingerprint"))
	viper.BindPFlag("ssh.only-failures", runCmd.Flags().Lookup("only-failures"))
	viper.BindPFlag("ssh.concurrency", runCmd.Flags().Lookup("concurrency"))
	viper.BindPFlag("ssh.use-private-ip", runCmd.Flags().Lookup("use-private-ip"))
	runCmd.SetUsageTemplate(usage)
}

func run(cmd *cobra.Command, args []string) error {
	command := strings.Join(args, " ")
	project := viper.GetString("gcp.projectid")
	timeout := viper.GetDuration("ssh.timeout")
	ignoreFingerprint := viper.GetBool("ssh.ignore-fingerprint")
	onlyFailures := viper.GetBool("ssh.only-failures")
	concurrency := viper.GetInt("ssh.concurrency")
	usePrivateIP := viper.GetBool("ssh.use-private-ip")
	useOSlogin := viper.GetBool("gcp.use-oslogin")

	target, err := cmd.Flags().GetString("target")
	if err != nil {
		return err
	}

	gcpClient, err := cloud.NewGCPClient(project)
	if err != nil {
		return err
	}

	path, err := key.Path()
	if err != nil {
		return err
	}

	k, err := key.Read(path)
	if err != nil {
		return err
	}

	log.Info("Fetching instance list")
	instances, err := gcpClient.GetInstances(target, usePrivateIP)
	if err != nil {
		return err
	}

	if len(instances) == 0 {
		log.Warn("No instances found")
		return nil
	}

	if useOSlogin {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		user, err := gcpClient.GetSAUsername(ctx)
		if err != nil {
			return err
		}
		k.User = user
	}

	m := ssh.NewMarathon(command, timeout, concurrency)
	if ignoreFingerprint {
		err = m.RunInsecure(instances, k)
		if err != nil {
			return err
		}
	} else {
		err = m.Run(instances, k)
		if err != nil {
			return err
		}
	}

	m.PrintResult(onlyFailures)
	return nil
}
