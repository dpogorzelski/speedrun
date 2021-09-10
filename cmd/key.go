package cmd

import (
	"context"

	"github.com/speedrunsh/speedrun/cloud"
	"github.com/speedrunsh/speedrun/key"

	"github.com/apex/log"
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

var listKeysCmd = &cobra.Command{
	Use:     "list",
	Short:   "List OS Login keys",
	Example: "  speedrun key list",
	PreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
	RunE: listKeys,
}

func init() {
	keyCmd.AddCommand(newKeyCmd)
	keyCmd.AddCommand(authorizeKeyCmd)
	keyCmd.AddCommand(revokeKeyCmd)
	keyCmd.AddCommand(listKeysCmd)
	authorizeKeyCmd.Flags().Bool("use-oslogin", false, "Authorize the key via OS Login rather than metadata")
	viper.BindPFlag("gcp.use-oslogin", authorizeKeyCmd.Flags().Lookup("use-oslogin"))
	keyCmd.SetUsageTemplate(usage)
}

func newKey(cmd *cobra.Command, args []string) error {
	k, err := key.New()
	if err != nil {
		return err
	}

	path, err := key.Path()
	if err != nil {
		return err
	}

	err = k.Write(path)
	if err != nil {
		return err
	}
	log.Info("Private key created")

	return nil
}

func authorizeKey(cmd *cobra.Command, args []string) error {
	project := viper.GetString("gcp.projectid")
	useOSlogin := viper.GetBool("gcp.use-oslogin")

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

	if useOSlogin {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		gcpClient.AddUserKey(ctx, k)
		if err != nil {
			return err
		}
		log.Info("Authorized key via OS Login")
	} else {
		gcpClient.AddKeyToMetadata(k)
		if err != nil {
			return err
		}
		log.Info("Authorized key in the project metadata")
	}

	return nil
}

func revokeKey(cmd *cobra.Command, args []string) error {
	project := viper.GetString("gcp.projectid")
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

	log.Info("Revoking public key")
	err = gcpClient.RemoveKeyFromMetadata(k)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = gcpClient.RemoveUserKey(ctx, k)
	if err != nil {
		return err
	}

	return nil
}

func listKeys(cmd *cobra.Command, args []string) error {
	project := viper.GetString("gcp.projectid")
	gcpClient, err := cloud.NewGCPClient(project)
	if err != nil {
		return err
	}

	log.Info("Fetching OS Login keys")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = gcpClient.ListUserKeys(ctx)
	if err != nil {
		return err
	}

	return nil
}
