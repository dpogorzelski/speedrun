package ssh

import (
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/apex/log"
	"github.com/melbahja/goph"
	"github.com/mitchellh/go-homedir"
	"github.com/speedrunsh/speedrun/key"
	"golang.org/x/crypto/ssh"
)

// VerifyHost chekcks that the remote host's fingerprint matches the know one to avoid MITM.
// If the host is new the fingerprint is added to known hostss file
func verifyHost(host string, remote net.Addr, key ssh.PublicKey) error {
	home, err := homedir.Dir()
	if err != nil {
		return err
	}

	knownhosts := filepath.Join(home, ".speedrun", "known_hosts")

	hostFound, err := goph.CheckKnownHost(host, remote, key, knownhosts)
	if hostFound && err != nil {
		log.Debugf("Host fingerprint known")
		return err
	}

	if !hostFound && err != nil {
		if err.Error() == "knownhosts: key is unknown" {
			log.Debugf("Adding host %s to ~/.speedrun/known_hosts", host)
			return goph.AddKnownHost(host, remote, key, knownhosts)
		}
		return err
	}

	if hostFound {
		log.Debugf("Host %s is already known", host)
		return nil
	}

	return nil
}

func checkHostsFile() error {
	home, err := homedir.Dir()
	if err != nil {
		return err
	}

	knownhosts := filepath.Join(home, ".speedrun", "known_hosts")

	if _, err := os.Stat(knownhosts); os.IsNotExist(err) {
		_, err = os.Create(knownhosts)
		if err != nil {
			return err
		}
	}
	return nil
}

func Connect(address string, key *key.Key) (*goph.Client, error) {
	auth, err := key.GetAuth()
	if err != nil {
		return nil, err
	}

	err = checkHostsFile()
	if err != nil {
		return nil, err
	}

	client, err := goph.NewConn(&goph.Config{
		User:     key.User,
		Addr:     address,
		Port:     22,
		Auth:     auth,
		Callback: verifyHost,
		Timeout:  time.Second * 10,
	})

	if err != nil {
		return nil, err
	}

	return client, nil
}

func ConnectInsecure(address string, key key.Key) (*goph.Client, error) {
	auth, err := key.GetAuth()
	if err != nil {
		return nil, err
	}

	client, err := goph.NewConn(&goph.Config{
		User:     key.User,
		Addr:     address,
		Port:     22,
		Auth:     auth,
		Callback: ssh.InsecureIgnoreHostKey(),
		Timeout:  time.Second * 10,
	})

	if err != nil {
		return nil, err
	}

	return client, nil
}
