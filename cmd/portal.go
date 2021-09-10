package cmd

import (
	"github.com/melbahja/goph"
	"github.com/speedrunsh/speedrun/cloud"
	"github.com/speedrunsh/speedrun/key"
	"github.com/speedrunsh/speedrun/ssh"

	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var portalCmd = &cobra.Command{
	Use:              "portal",
	Short:            "Manage portals",
	TraverseChildren: true,
}

var deployCmd = &cobra.Command{
	Use:     "deploy",
	Short:   "Deploy a portal on remote servers",
	Example: "  speedrun portal deploy",
	PreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
	RunE: deploy,
}

func init() {
	portalCmd.PersistentFlags().StringP("target", "t", "", "Select instances that match the given criteria")
	portalCmd.PersistentFlags().Bool("ignore-fingerprint", false, "Ignore host's fingerprint mismatch")
	viper.BindPFlag("ssh.ignore-fingerprint", portalCmd.PersistentFlags().Lookup("ignore-fingerprint"))
	portalCmd.AddCommand(deployCmd)
	portalCmd.SetUsageTemplate(usage)
}

func deploy(cmd *cobra.Command, args []string) error {
	project := viper.GetString("gcp.projectid")
	target, err := cmd.Flags().GetString("target")
	ignoreFingerprint := viper.GetBool("ssh.ignore-fingerprint")
	if err != nil {
		return err
	}

	gcpClient, err := cloud.NewGCPClient(project)
	if err != nil {
		return err
	}

	log.Info("Fetching instance list")
	instances, err := gcpClient.GetInstances(target, false)
	if err != nil {
		return err
	}

	if len(instances) == 0 {
		log.Warn("No instances found")
		return nil
	}

	path, err := key.Path()
	if err != nil {
		return err
	}

	k, err := key.Read(path)
	if err != nil {
		return err
	}

	for _, instance := range instances {
		var sshclient *goph.Client
		if ignoreFingerprint {
			sshclient, err = ssh.ConnectInsecure(instance.Address, k)
		} else {
			sshclient, err = ssh.Connect(instance.Address, k)
		}
		if err != nil {
			log.WithField("instance", instance.Name).Errorf("failed to establish SSH connection: %v", err)
			continue
		}
		err = sshclient.Upload("./tpl", "tmp/tpl")
		if err != nil {
			log.WithField("instance", instance.Name).Errorf("failed to deploy portal: %v", err)
			continue
		}
	}
	return nil
}
