package ssh

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/speedrunsh/speedrun/cloud"
	"github.com/speedrunsh/speedrun/colors"
	"github.com/speedrunsh/speedrun/key"

	"github.com/alitto/pond"
	"github.com/apex/log"
	"github.com/cheggaaa/pb/v3"
	"github.com/melbahja/goph"
	"github.com/spf13/viper"
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
func NewMarathon(command string, timeout time.Duration, concurrency int) *Marathon {
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
func (m *Marathon) Run(instances []cloud.Instance, key *key.Key) error {
	auth, err := key.GetAuth()
	if err != nil {
		return err
	}

	err = checkHostsFile()
	if err != nil {
		return err
	}

	pool := pond.New(m.Concurrency, 10000)

	bar := pb.New(len(instances))
	lvl, err := log.ParseLevel(viper.GetString("loglevel"))
	if err != nil {
		return fmt.Errorf("couldn't parse log level: %s", err)
	}

	if lvl > 0 {
		bar.SetMaxWidth(1)
		bar.SetTemplateString(fmt.Sprintf("%s Running [%s]: {{counters . }}", colors.Blue("•"), colors.Purple(m.Command)))
		bar.Start()
	}

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
				Callback: verifyHost,
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

func (m *Marathon) RunInsecure(instances []cloud.Instance, key *key.Key) error {
	auth, err := key.GetAuth()
	if err != nil {
		return err
	}

	pool := pond.New(m.Concurrency, 10000)

	bar := pb.New(len(instances))
	lvl, err := log.ParseLevel(viper.GetString("loglevel"))
	if err != nil {
		return fmt.Errorf("couldn't parse log level: %s", err)
	}

	if lvl > 0 {
		bar.SetMaxWidth(1)
		bar.SetTemplateString(fmt.Sprintf("%s Running [%s]: {{counters . }}", colors.Blue("•"), colors.Purple(m.Command)))
		bar.Start()
	}

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
				Callback: ssh.InsecureIgnoreHostKey(),
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

func formatOutput(body string) string {
	f := []string{}
	for _, line := range strings.Split(body, "\n") {
		line = fmt.Sprintf("    %s", line)
		f = append(f, line)
	}

	f = append(f[:len(f)-1], "")
	return strings.Join(f, "\n")
}
