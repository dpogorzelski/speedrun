package cmd

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"speedrun/cloud"

	"github.com/apex/log"
	"github.com/mikesmitty/edkey"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
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
	log.Debug("Generating new private key")
	_, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	pemKey := &pem.Block{
		Type:  "OPENSSH PRIVATE KEY",
		Bytes: edkey.MarshalED25519PrivateKey(privKey),
	}
	privateKey := pem.EncodeToMemory(pemKey)

	err = writeKeyFile(privateKey)
	if err != nil {
		return err
	}

	return nil
}

func writeKeyFile(key []byte) error {
	privateKeyPath, err := determineKeyFilePath()
	if err != nil {
		return err
	}

	log.Debugf("Writing priviate key to %s", privateKeyPath)
	err = ioutil.WriteFile(privateKeyPath, key, 0600)
	if err != nil {
		return err
	}
	return nil
}

func loadPubKey() ([]byte, error) {
	privateKeyPath, err := determineKeyFilePath()
	if err != nil {
		return nil, err
	}

	file, err := readKeyFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("couldn't find private key. Use 'speedrun key new' to generate a new one")
	}

	privKey, err := ssh.ParsePrivateKey(file)
	if err != nil {
		return nil, err
	}

	authorizedKey := ssh.MarshalAuthorizedKey(privKey.PublicKey())
	return authorizedKey, nil
}

func readKeyFile(path string) ([]byte, error) {
	cleanPath := filepath.Clean(path)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return nil, err
	}

	file, err := ioutil.ReadFile(absPath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func authorizeKey(cmd *cobra.Command, args []string) error {
	project := viper.GetString("gcp.projectid")
	client, err := cloud.NewClient(cloud.SetProject(project))
	if err != nil {
		return err
	}

	pubKey, err := loadPubKey()
	if err != nil {
		return err
	}

	log.Infof("Auhtorizing public key: %s", pubKey)

	client.AuthorizeKey(pubKey)
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

	pubKey, err := loadPubKey()
	if err != nil {
		return err
	}

	log.Info("Revoking public key")
	err = client.RevokeKey(pubKey)
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
