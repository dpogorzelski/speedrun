package cmd

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"path/filepath"

	gcp "speedrun/cloud"

	"github.com/alitto/pond"
	"github.com/apex/log"
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

var setKeyCmd = &cobra.Command{
	Use:     "set",
	Short:   "Set key in the project or instance metadata",
	Example: "  speedrun key set \n  speedrun key set --filter \"labels.foo = bar AND labels.environment = staging\"",
	PreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
	RunE: setKey,
}

var removeKeyCmd = &cobra.Command{
	Use:     "remove",
	Short:   "Remove key from the project metadata or instance metadata",
	Example: "  speedrun key remove \n  speedrun key remove --filter \"labels.foo = bar AND labels.environment = staging\"",
	PreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
	RunE: removeKey,
}

func init() {
	setKeyCmd.Flags().String("filter", "", "Set the key only on matching instances")
	removeKeyCmd.Flags().String("filter", "", "Set the key only on matching instances")
	keyCmd.AddCommand(newKeyCmd)
	keyCmd.AddCommand(setKeyCmd)
	keyCmd.AddCommand(removeKeyCmd)
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

	log.Debug("Converting private key to PKCS8 format")
	pemBlock := &pem.Block{}
	pemBlock.Type = "PRIVATE KEY"
	pemBlock.Bytes, err = x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return err
	}

	log.Debug("Encoding the key to PEM format")
	privateKey := pem.EncodeToMemory(pemBlock)

	err = writeKeyFile(privateKey)
	if err != nil {
		return err
	}
	log.Info("Private key created")

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

func loadKeyPair() (ssh.PublicKey, ssh.Signer, error) {
	privateKeyPath, err := determineKeyFilePath()
	if err != nil {
		return nil, nil, err
	}

	file, err := readKeyFile(privateKeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't find private key. Use 'speedrun key new' to generate a new one")
	}

	signer, err := ssh.ParsePrivateKey(file)
	if err != nil {
		return nil, nil, err
	}
	pubKey := signer.PublicKey()
	return pubKey, signer, nil
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

func setKey(cmd *cobra.Command, args []string) error {
	client, err := gcp.NewComputeClient(viper.GetString("gcp.projectid"))
	if err != nil {
		return err
	}

	pubKey, _, err := loadKeyPair()
	if err != nil {
		return err
	}

	filter, err := cmd.Flags().GetString("filter")
	if err != nil {
		return err
	}

	if filter != "" {
		log.Info("Setting public key in the instance metadata")
		instances, err := client.GetInstances(filter)
		if err != nil {
			return err
		}

		if len(instances) == 0 {
			log.Warn("no instances found")
		}

		pool := pond.New(10, 0, pond.MinWorkers(10))
		for i := 0; i < len(instances); i++ {
			n := i
			pool.Submit(func() {
				client.AddKeyToMetadata(instances[n], pubKey)
			})
		}

		pool.StopAndWait()
	} else {
		log.Info("Setting public key in the project metadata")
		err = client.AddKeyToMetadataP(pubKey)
		if err != nil {
			return err
		}
	}

	return nil
}

func removeKey(cmd *cobra.Command, args []string) error {
	client, err := gcp.NewComputeClient(viper.GetString("gcp.projectid"))
	if err != nil {
		return err
	}

	pubKey, _, err := loadKeyPair()
	if err != nil {
		return err
	}

	filter, err := cmd.Flags().GetString("filter")
	if err != nil {
		return err
	}

	if filter != "" {
		log.Info("Removing public from the instance metadata")
		instances, err := client.GetInstances(filter)
		if err != nil {
			return err
		}

		if len(instances) == 0 {
			log.Warn("no instances found")
		}

		pool := pond.New(10, 0, pond.MinWorkers(10))
		for i := 0; i < len(instances); i++ {
			n := i
			pool.Submit(func() {
				client.RemoveKeyFromMetadata(instances[n], pubKey)
			})
		}

		pool.StopAndWait()
	} else {
		log.Info("Removing public key from the project metadata")
		err = client.RemoveKeyFromMetadataP(pubKey)
		if err != nil {
			return err
		}
	}

	return nil
}
