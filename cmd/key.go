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

var listKeyCmd = &cobra.Command{
	Use:     "list",
	Short:   "List user keys",
	Example: "  speedrun key list",
	PreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
	RunE: listKey,
}

func init() {
	keyCmd.AddCommand(newKeyCmd)
	keyCmd.AddCommand(authorizeKeyCmd)
	keyCmd.AddCommand(revokeKeyCmd)
	keyCmd.AddCommand(listKeyCmd)
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
	client, err := cloud.NewClient(cloud.SetProject(project))
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
	client.AuthorizeKey(k)
	if err != nil {
		return err
	}

	return nil
}

func revokeKey(cmd *cobra.Command, args []string) error {
	project := viper.GetString("gcp.projectid")
	client, err := cloud.NewClient(cloud.SetProject(project))
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
	err = client.RevokeKey(k)
	if err != nil {
		return err
	}

	return nil
}

func listKey(cmd *cobra.Command, args []string) error {
	project := viper.GetString("gcp.projectid")
	client, err := cloud.NewClient(cloud.SetProject(project))
	if err != nil {
		return err
	}

	err = client.ListKeys()
	if err != nil {
		return err
	}

	return nil
}
