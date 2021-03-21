package marathon

import (
	"fmt"
	"net"
	"os/user"
	"speedrun/colors"
	"sync"
	"time"

	"github.com/alitto/pond"
	"github.com/apex/log"
	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"
)

// Marathon represents the instance of the execution of a command against a number of target servers
type Marathon struct {
	sync.Mutex
	errors      map[string]error
	failures    map[string]string
	successes   map[string]string
	Command     string
	Timeout     time.Duration
	Concurrency int
}

// New creates a new instance of the Marathon type
func New(command string, timeout time.Duration, concurrency int) *Marathon {
	r := Marathon{
		errors:      make(map[string]error),
		failures:    make(map[string]string),
		successes:   make(map[string]string),
		Command:     command,
		Timeout:     timeout,
		Concurrency: concurrency,
	}

	return &r
}

// Run runs a given command on servers in the addresses list
func (m *Marathon) Run(instances map[string]string, key string, ignoreFingerprint bool) error {
	user, err := user.Current()
	if err != nil {
		return err
	}

	auth, err := goph.Key(key, "")
	if err != nil {
		return err
	}

	cb := verifyHost
	if ignoreFingerprint {
		cb = ssh.InsecureIgnoreHostKey()
	}

	pool := pond.New(m.Concurrency, 10000)

	for k, v := range instances {
		addr := k
		host := v
		log.Debugf("Adding %s(%s) to the queue", v, k)
		pool.Submit(func() {
			var client *goph.Client
			var err error

			for i := 0; i < 60; i++ {
				log.WithField("host", host).Debug("Checking if they public key has propagated yet")
				client, err = goph.NewConn(&goph.Config{
					User:     user.Username,
					Addr:     addr,
					Port:     22,
					Auth:     auth,
					Callback: cb,
					Timeout:  m.Timeout,
				})
				if err != nil && err.Error() == "ssh: handshake failed: ssh: unable to authenticate, attempted methods [none publickey], no supported methods remain" {
					time.Sleep(time.Second)
				} else {
					break
				}
			}

			if err != nil {
				log.WithField("host", host).Debugf("Error encountered while trying to connect: %s", err)
				m.Lock()
				m.errors[host] = err
				m.Unlock()
				return
			}
			defer client.Close()

			out, err := client.Run(m.Command)
			if err != nil {
				m.Lock()
				m.failures[host] = formatOutput(string(out))
				m.Unlock()
				return
			}
			m.Lock()
			m.successes[host] = formatOutput(string(out))
			m.Unlock()
		})
	}
	pool.StopAndWait()

	return nil
}

// VerifyHost chekcks that the remote host's fingerprint matches the know one to avoid MITM.
// If the host is new the fingerprint is added to known hostss file
func verifyHost(host string, remote net.Addr, key ssh.PublicKey) error {
	hostFound, err := goph.CheckKnownHost(host, remote, key, "")
	if hostFound && err != nil {
		log.Debugf("Host fingerprint known")
		return err
	}

	if !hostFound && err != nil {
		if err.Error() == "knownhosts: key is unknown" {
			log.Warnf("Adding host %s to ~/.ssh/known_hosts", host)
			return goph.AddKnownHost(host, remote, key, "")
		}
		return err
	}

	if hostFound {
		log.Debugf("Host %s is already known", host)
		return nil
	}

	return nil
}

// PrintResult prints the results of the ssh command run
func (m *Marathon) PrintResult(failures bool) {

	if !failures {
		for host, msg := range m.successes {
			fmt.Printf("  %s:\n%s\n", colors.Green(host), colors.White(msg))
		}
	}

	for host, msg := range m.failures {
		fmt.Printf("  %s:\n%s\n", colors.Yellow(host), colors.White(msg))
	}

	for host, msg := range m.errors {
		fmt.Printf("  %s:\n    %s\n\n", colors.Red(host), colors.White(msg.Error()))
	}
	fmt.Printf("%s: %d\n%s: %d\n%s:   %d\n", colors.Green("Success"), len(m.successes), colors.Yellow("Failure"), len(m.failures), colors.Red("Error"), len(m.errors))
}
