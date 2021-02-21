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
	"github.com/fatih/color"
	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"
)

// Marathon represents the instance of the execution of a command against a number of target servers
type Marathon struct {
	sync.Mutex
	errors    map[string]error
	failures  map[string]string
	successes map[string]string
	Command   string
	Timeout   time.Duration
}

// New creates a new instance of the Marathon type
func New(command string, timeout time.Duration) *Marathon {
	r := Marathon{
		errors:    make(map[string]error),
		failures:  make(map[string]string),
		successes: make(map[string]string),
		Command:   command,
		Timeout:   timeout,
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

	pool := pond.New(100, 0)

	for k, v := range instances {
		addr := k
		host := v
		pool.Submit(func() {
			client, err := goph.NewConn(&goph.Config{
				User:     user.Username,
				Addr:     addr,
				Port:     22,
				Auth:     auth,
				Callback: cb,
				Timeout:  m.Timeout,
			})
			if err != nil {
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
	if err != nil {
		return err
	}

	if hostFound {
		log.Debugf("Host %s is already known", host)
		return nil
	}

	log.Debugf("Adding host %s to ~/.ssh/known_hosts", host)
	return goph.AddKnownHost(host, remote, key, "")
}

// PrintResult prints the results of the ssh command run
func (m *Marathon) PrintResult(failures bool) {

	output := color.New(color.FgWhite).SprintFunc()

	if !failures {
		for host, msg := range m.successes {
			fmt.Printf("  %s:\n%s\n", colors.Green(host), output(msg))
		}
	}

	for host, msg := range m.failures {
		fmt.Printf("  %s:\n%s\n", colors.Yellow(host), output(msg))
	}

	for host, msg := range m.errors {
		fmt.Printf("  %s:\n    %s\n\n", colors.Red(host), output(msg.Error()))
	}
	fmt.Printf("%s: %d %s: %d %s: %d\n", colors.Green("Success"), len(m.successes), colors.Yellow("Failure"), len(m.failures), colors.Red("Error"), len(m.errors))
}
