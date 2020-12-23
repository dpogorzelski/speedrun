package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh"
)

// type key struct {
// 	privatekey []byte
// }

// func newkey() (key, error) {
// 	key := new(key)

// 	return key, nil
// }

// func (k *key) create(ctx *cli.Context) error {
// 	_, privKey, err := ed25519.GenerateKey(rand.Reader)
// 	if err != nil {
// 		return cli.Exit(err, 1)
// 	}

// 	pemBlock := &pem.Block{}
// 	pemBlock.Type = "PRIVATE KEY"
// 	pemBlock.Bytes, err = x509.MarshalPKCS8PrivateKey(privKey)
// 	if err != nil {
// 		return cli.Exit(err, 1)
// 	}

// 	privateKey := pem.EncodeToMemory(pemBlock)

// 	err = writeKeyFile(privateKey)
// 	if err != nil {
// 		return cli.Exit(err, 1)
// 	}

// 	fmt.Println("generated new ssh key")
// 	return nil
// }

// func (k *key) writeToFile() error {
// 	privateKeyPath, err := determineKeyFilePath()
// 	if err != nil {
// 		return err
// 	}

// 	err = ioutil.WriteFile(privateKeyPath, k.privatekey, 0600)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func determineKeyFilePath() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(home, ".config/speedrun/privatekey")
	return path, nil
}

func createKey(ctx *cli.Context) error {
	_, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return cli.Exit(err, 1)
	}

	pemBlock := &pem.Block{}
	pemBlock.Type = "PRIVATE KEY"
	pemBlock.Bytes, err = x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return cli.Exit(err, 1)
	}

	privateKey := pem.EncodeToMemory(pemBlock)

	err = writeKeyFile(privateKey)
	if err != nil {
		return cli.Exit(err, 1)
	}

	fmt.Println("generated new ssh key")
	return nil
}

func writeKeyFile(key []byte) error {
	privateKeyPath, err := determineKeyFilePath()
	if err != nil {
		return err
	}

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

func showKey(ctx *cli.Context) error {
	fmt.Println("showing private key")
	return nil
}
