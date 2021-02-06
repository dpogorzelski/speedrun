package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/alitto/pond"
	"github.com/apex/log"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh"
)

func determineKeyFilePath() (string, error) {
	log.Debug("Determining private key path")
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(home, ".speedrun/privatekey")
	return path, nil
}

func createKey(c *cli.Context) error {
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
		return nil, nil, fmt.Errorf("Couldn't find private key. Use 'speedrun key new' to generate a new one")
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

func setKey(c *cli.Context) error {
	client, err := NewComputeClient(config.Gcp.Projectid)
	if err != nil {
		return cli.Exit(err, 1)
	}

	pubKey, _, err := loadKeyPair()
	if err != nil {
		return cli.Exit(err, 1)
	}

	p := NewProgress()
	p.Start("Setting public key in the project metadata")
	err = client.UpdateProjectMetadata(pubKey)
	if err != nil {
		p.Error(err)
	}
	p.Stop()

	filter := c.String("filter")
	p.Start("Setting public key in the instance metadata")
	instances, err := client.GetInstances(filter)
	if err != nil {
		p.Error(err)
	}

	if len(instances) == 0 {
		p.Failure("no instances found")
	}

	pool := pond.New(10, 0, pond.MinWorkers(10))
	for i := 0; i < len(instances); i++ {
		n := i
		pool.Submit(func() {
			client.UpdateInstanceMetadata(instances[n], pubKey)
		})
	}

	pool.StopAndWait()
	p.Stop()
	return nil
}
