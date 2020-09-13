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

	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/yahoo/vssh"
	"golang.org/x/crypto/ssh"
	"google.golang.org/api/compute/v1"
)

type result struct {
	errors    map[string]error
	failures  map[string]string
	successes map[string]string
}

type instanceList map[string]string

type Run struct {
	res       result
	instances instanceList
}

const (
	GREEN = iota
	YELLOW
	RED
)

func NewRun() *Run {
	r := Run{}
	r.res.errors = make(map[string]error)
	r.res.failures = make(map[string]string)
	r.res.successes = make(map[string]string)
	return &r
}

// GetKeyPair returns a public key, either new or existing depending on the force bool value. The key is formatted for use in authorized_keys files or GCP metadata.
func GetKeyPair(force bool) (ssh.PublicKey, ssh.Signer, error) {
	var sshPubKey ssh.PublicKey
	var signer ssh.Signer
	var err error

	if force {
		sshPubKey, signer, err = GenerateKeyPair()
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

func GenerateKeyPair() (ssh.PublicKey, ssh.Signer, error) {
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
		return nil, nil, fmt.Errorf("Couldn't load private key from the expected location. Use --force-new-key to generate a new one: %v", err)
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
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	fileName, err := determineKeyFileName()
	if err != nil {
		return "", err
	}
	path := filepath.Join(home, ".nyx", fileName)
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
func Execute(command string, instances []*compute.Instance, key ssh.Signer) (*Run, error) {
	vs := vssh.New().Start()
	user, err := user.Current()
	run := NewRun()
	if err != nil {
		return run, err
	}

	config, err := getSSHConfig(user.Username, key)
	if err != nil {
		return run, err
	}

	instanceDict := map[string]string{}
	for _, instance := range instances {
		instanceDict[instance.NetworkInterfaces[0].AccessConfigs[0].NatIP+":22"] = instance.Name
	}

	for addr := range instanceDict {
		err := vs.AddClient(addr, config, vssh.SetMaxSessions(10))
		if err != nil {
			return run, err
		}
	}
	vs.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	timeout, err := time.ParseDuration("20s")
	if err != nil {
		return run, err
	}

	respChan := vs.Run(ctx, command, timeout, vssh.SetLimitReaderStdout(4096))

	for resp := range respChan {
		if err := resp.Err(); err != nil {
			run.res.errors[instanceDict[resp.ID()]] = err
			continue
		}

		output, _, _ := resp.GetText(vs)
		if resp.ExitStatus() == 0 {
			run.res.successes[instanceDict[resp.ID()]] = output
		} else {
			run.res.failures[instanceDict[resp.ID()]] = output
		}

	}
	return run, nil
}

// PrintResult prints the results of the ssh command run
func (r Run) PrintResult(failures bool) {
	green := color.New(color.FgGreen).SprintfFunc()
	yellow := color.New(color.FgYellow).SprintfFunc()
	red := color.New(color.FgRed).SprintfFunc()

	if failures {
		for k, v := range r.res.successes {
			fmt.Printf("%s:\n%s\n", green(k), v)
		}
	}

	for k, v := range r.res.failures {
		fmt.Printf("%s:\n%s\n", yellow(k), v)
	}

	for k, v := range r.res.errors {
		fmt.Printf("%s:\n%s\n\n", red(k), v.Error())
	}
	fmt.Printf("%s: %d %s: %d %s: %d\n", green("Success"), len(r.res.successes), yellow("Failure"), len(r.res.failures), red("Error"), len(r.res.errors))
}

// Status function returns true if there are no errors or failures in a run, false otherwise
func (r Run) Status() interface{} {
	if len(r.res.errors) > 0 {
		return RED
	}

	if len(r.res.failures) > 0 {
		return YELLOW
	}
	return GREEN
}

func padOutput(body string) string {
	f := []string{}
	for _, line := range strings.Split(body, "\n") {
		line = fmt.Sprintf("  %s", line)
		f = append(f, line)
	}
	return strings.Join(f, "\n")
}
