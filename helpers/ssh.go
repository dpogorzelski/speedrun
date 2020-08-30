package helpers

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/yahoo/vssh"
	"golang.org/x/crypto/ssh"
)

// GetKeyPair returns a public key, either new or existing depending on the force bool value. The key is formatted for use in authorized_keys files or GCP metadata.
func GetKeyPair(force bool) (ssh.PublicKey, ssh.Signer, error) {
	var sshPubKey ssh.PublicKey
	var signer ssh.Signer
	var err error

	if force {
		sshPubKey, signer, err = generateKeyPair()
		if err != nil {
			return nil, nil, err
		}
	} else {
		sshPubKey, signer, err = loadKeyPair()
		if err != nil {
			return nil, nil, err
		}
	}
	return sshPubKey, signer, nil
}

func generateKeyPair() (ssh.PublicKey, ssh.Signer, error) {
	_, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	pemBlock := &pem.Block{}
	pemBlock.Type = "PRIVATE KEY"
	pemBlock.Bytes, err = x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return nil, nil, err
	}

	privateKey := pem.EncodeToMemory(pemBlock)

	err = writeKeyFile(privateKey)
	if err != nil {
		return nil, nil, err
	}

	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, nil, err
	}
	pubKey := signer.PublicKey()
	return pubKey, signer, nil
}

func loadKeyPair() (ssh.PublicKey, ssh.Signer, error) {
	privateKeyPath, err := determineKeyFilePath()
	if err != nil {
		return nil, nil, err
	}

	file, err := readKeyFile(privateKeyPath)
	if err != nil {
		return nil, nil, err
	}

	log.Debugf("Parsing private key")
	signer, err := ssh.ParsePrivateKey(file)
	if err != nil {
		return nil, nil, err
	}
	pubKey := signer.PublicKey()
	return pubKey, signer, nil
}

func determineKeyFileName() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(user.Username))
	name := fmt.Sprintf("%x", sum)
	log.Debugf("Determined private key file name %s", name)
	return name, nil
}

func determineKeyFilePath() (string, error) {
	homeDir, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	fileName, err := determineKeyFileName()
	if err != nil {
		return "", err
	}
	path := filepath.Join(homeDir, ".nyx", fileName)
	log.Debugf("Determined private key file path %s", path)
	return path, nil
}

func readKeyFile(path string) ([]byte, error) {
	cleanPath := filepath.Clean(path)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return nil, err
	}
	log.Debugf("Reading private key from %s", absPath)
	file, err := ioutil.ReadFile(absPath)
	if err != nil {
		return nil, err
	}
	return file, nil
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

func getSSHConfig(user string, key ssh.Signer) (*ssh.ClientConfig, error) {
	var auths []ssh.AuthMethod
	auths = append(auths, ssh.PublicKeys(key))
	return &ssh.ClientConfig{
		User:            user,
		Auth:            auths,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil
}

// Execute runs agiven command on servers in the addresses list
func Execute(command string, addresses []string, key ssh.Signer) error {
	vs := vssh.New().Start()
	user, err := user.Current()
	if err != nil {
		return err
	}

	config, err := getSSHConfig(user.Username, key)
	if err != nil {
		return err
	}
	for _, addr := range addresses {
		err := vs.AddClient(addr, config, vssh.SetMaxSessions(10))
		if err != nil {
			return err
		}
	}
	vs.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	timeout, err := time.ParseDuration("10s")
	if err != nil {
		return err
	}
	respChan := vs.Run(ctx, command, timeout)

	for resp := range respChan {
		if err := resp.Err(); err != nil {
			log.WithFields(logrus.Fields{
				"prefix": resp.ID(),
			}).Errorln(err)
			continue
		}
		outTxt, _, _ := resp.GetText(vs)
		outTxt = padOutput(outTxt)
		if resp.ExitStatus() == 0 {
			log.WithFields(logrus.Fields{
				"prefix": resp.ID(),
			}).Infof("\n%s", outTxt)
		} else {
			log.WithFields(logrus.Fields{
				"prefix": resp.ID(),
			}).Errorf("\n%s", outTxt)
		}

	}
	return nil
}

func padOutput(body string) string {
	f := []string{}
	for _, line := range strings.Split(body, "\n") {
		line = fmt.Sprintf("  %s", line)
		f = append(f, line)
	}
	return strings.Join(f, "\n")
}
