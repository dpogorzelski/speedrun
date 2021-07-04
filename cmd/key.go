package cmd

import (
	"path/filepath"

	"speedrun/cloud"
	"speedrun/key"

	"github.com/apex/log"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var keyCmd = &cobra.Command{
	Use:              "key",
	Short:            "Manage ssh keys",
	TraverseChildren: true,
}

var newKeyCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new ssh key",
	PreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
	RunE: newKey,
}

var authorizeKeyCmd = &cobra.Command{
	Use:     "authorize",
	Short:   "Authorize key for ssh access",
	Example: "  speedrun key authorize",
	PreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
	RunE: authorizeKey,
}

var revokeKeyCmd = &cobra.Command{
	Use:     "revoke",
	Short:   "Revoke ssh key",
	Example: "  speedrun key revoke",
	PreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
	RunE: revokeKey,
}

func init() {
	keyCmd.AddCommand(newKeyCmd)
	keyCmd.AddCommand(authorizeKeyCmd)
	keyCmd.AddCommand(revokeKeyCmd)
	authorizeKeyCmd.Flags().Bool("use-oslogin", false, "Authorize the key via OS Login rather than metadata")
	viper.BindPFlag("gcp.use-oslogin", authorizeKeyCmd.Flags().Lookup("use-oslogin"))
}

func determineKeyFilePath() (string, error) {
	log.Debug("Determining private key path")
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(home, ".speedrun/privatekey")
	return path, nil
}

func newKey(cmd *cobra.Command, args []string) error {
	k, err := key.New()
	if err != nil {
		return err
	}

	path, err := determineKeyFilePath()
	if err != nil {
		return err
	}

	err = k.Write(path)
	if err != nil {
		return err
	}

	return nil
}

func authorizeKey(cmd *cobra.Command, args []string) error {
	project := viper.GetString("gcp.projectid")
	useOSlogin := viper.GetBool("gcp.use-oslogin")

	gcpClient, err := cloud.NewGCPClient(project)
	if err != nil {
		return err
	}

	path, err := determineKeyFilePath()
	if err != nil {
		return err
	}

	k, err := key.Read(path)
	if err != nil {
		return err
	}

	log.Infof("Authorizing public key")
	if useOSlogin {
		gcpClient.AddUserKey(k)
		if err != nil {
			return err
		}
	} else {
		gcpClient.AddKeyToMetadata(k)
		if err != nil {
			return err
		}
	}

	return nil
}

func revokeKey(cmd *cobra.Command, args []string) error {
	project := viper.GetString("gcp.projectid")
	gcpClient, err := cloud.NewGCPClient(project)
	if err != nil {
		return err
	}

	path, err := determineKeyFilePath()
	if err != nil {
		return err
	}

	k, err := key.Read(path)
	if err != nil {
		return err
	}

	log.Info("Revoking public key")
	err = gcpClient.RemoveKeyFromMetadata(k)
	if err != nil {
		return err
	}

	err = gcpClient.RemoveUserKey(k)
	if err != nil {
		return err
	}

	return nil
}
