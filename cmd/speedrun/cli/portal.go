package cli

import (
	"fmt"

	"github.com/melbahja/goph"
	"github.com/speedrunsh/speedrun/pkg/common/key"
	"github.com/speedrunsh/speedrun/pkg/common/ssh"
	"github.com/speedrunsh/speedrun/pkg/speedrun/cloud"

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
	deployCmd.Flags().Bool("upload", false, "Upload portal's binary via local machine")
	deployCmd.Flags().Bool("install", true, "Install portal as a service")
	viper.BindPFlag("ssh.ignore-fingerprint", portalCmd.PersistentFlags().Lookup("ignore-fingerprint"))
	portalCmd.AddCommand(deployCmd)
	portalCmd.SetUsageTemplate(usage)
}

func deploy(cmd *cobra.Command, args []string) error {
	project := viper.GetString("gcp.projectid")
	target, err := cmd.Flags().GetString("target")
	if err != nil {
		return err
	}
	ignoreFingerprint := viper.GetBool("ssh.ignore-fingerprint")
	upload, err := cmd.Flags().GetBool("upload")
	if err != nil {
		return err
	}

	install, err := cmd.Flags().GetBool("install")
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

	if upload {
		log.Debug("will upload")
	}

	if install {
		log.Debug("will upload")
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
		// err = sshclient.Upload("./tpl", "tmp/tpl")
		arch, err := sshclient.Run("uname -m")
		if err != nil {
			log.WithField("instance", instance.Name).Errorf("failed to deploy portal: %v", err)
			continue
		}

		url, err := getUrl(string(arch))
		if err != nil {
			log.WithField("instance", instance.Name).Errorf("failed to deploy portal: %v", err)
			continue
		}

		downloadCommand := fmt.Sprintf("curl -L --silent %s -o /tmp/portal.zip", url)

		_, err = sshclient.Run(downloadCommand)
		if err != nil {
			log.WithField("instance", instance.Name).Errorf("failed to deploy portal: %v", err)
			continue
		}
	}
	return nil
}

func getUrl(arch string) (string, error) {
	var url string
	switch string(arch) {
	case "x86_64":
		url = "https://download.speedrun.sh/portal-linux-amd64.zip"
	case "armv8l":
		url = "https://download.speedrun.sh/portal-linux-arm64.zip"
	case "armv8b":
		url = "https://download.speedrun.sh/portal-linux-arm64.zip"
	case "aarch64":
		url = "https://download.speedrun.sh/portal-linux-arm64.zip"
	case "aarch64_be":
		url = "https://download.speedrun.sh/portal-linux-arm64.zip"
	default:
		return "", fmt.Errorf("unsupported CPU architecture")
	}
	return url, nil
}
