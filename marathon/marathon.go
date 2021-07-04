package marathon

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"speedrun/cloud"
	"speedrun/colors"
	"speedrun/key"
	"sync"
	"time"

	"github.com/alitto/pond"
	"github.com/apex/log"
	"github.com/cheggaaa/pb/v3"
	"github.com/melbahja/goph"
	"github.com/mitchellh/go-homedir"
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
func (m *Marathon) Run(instances []cloud.Instance, key *key.Key, ignoreFingerprint bool) error {
	auth, err := key.GetAuth()
	if err != nil {
		return err
	}

	err = checkHostsFile()
	if err != nil {
		return err
	}

	cb := verifyHost
	if ignoreFingerprint {
		cb = ssh.InsecureIgnoreHostKey()
	}

	pool := pond.New(m.Concurrency, 10000)

	bar := pb.New(len(instances))
	bar.SetMaxWidth(1)
	bar.SetTemplateString(fmt.Sprintf("%s Running [%s]: {{counters . }}", colors.Blue("•"), colors.Blue(m.Command)))
	bar.Start()

	for _, i := range instances {
		instance := i
		log.Debugf("Adding %s to the queue", instance.Name)
		pool.Submit(func() {
			var client *goph.Client
			var err error

			client, err = goph.NewConn(&goph.Config{
				User:     key.User,
				Addr:     instance.Address,
				Port:     22,
				Auth:     auth,
				Callback: cb,
				Timeout:  m.Timeout,
			})

			if err != nil {
				log.WithField("host", instance.Name).Debugf("Error encountered while trying to connect: %s", err)
				m.Lock()
				bar.Increment()
				m.errors[instance.Name] = err
				m.Unlock()
				return
			}
			defer client.Close()

			out, err := client.Run(m.Command)
			if err != nil {
				m.Lock()
				bar.Increment()
				m.failures[instance.Name] = formatOutput(string(out))
				m.Unlock()
				return
			}
			m.Lock()
			bar.Increment()
			m.successes[instance.Name] = formatOutput(string(out))
			m.Unlock()
		})
	}
	pool.StopAndWait()
	bar.Finish()

	return nil
}

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
			log.Warnf("Adding host %s to ~/.speedrun/known_hosts", host)
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
