package main

import (
	"fmt"
	"os/user"
	"strings"
	"sync"
	"time"

	"github.com/alitto/pond"
	"github.com/fatih/color"
	"github.com/melbahja/goph"
	"google.golang.org/api/compute/v1"
)

type roll struct {
	sync.Mutex
	errors    map[string]error
	failures  map[string]string
	successes map[string]string
	command   string
	timeout   time.Duration
}

func newRoll(command string, timeout time.Duration) *roll {
	r := roll{
		errors:    make(map[string]error),
		failures:  make(map[string]string),
		successes: make(map[string]string),
		command:   command,
		timeout:   timeout,
	}

	return &r
}

// Execute runs agiven command on servers in the addresses list
func (r *roll) execute(instances []*compute.Instance, key string) error {
	user, err := user.Current()
	if err != nil {
		return err
	}

	instanceDict := map[string]string{}
	for _, instance := range instances {
		instanceDict[instance.NetworkInterfaces[0].AccessConfigs[0].NatIP] = instance.Name
	}

	auth, err := goph.Key(key, "")
	if err != nil {
		return err
	}

	pool := pond.New(100, 0)

	for k, v := range instanceDict {
		addr := k
		host := v
		pool.Submit(func() {
			client, err := goph.NewUnknown(user.Username, addr, auth)
			client.Config.Timeout = r.timeout
			if err != nil {
				r.Lock()
				r.errors[host] = err
				r.Unlock()
				return
			}
			defer client.Close()
			out, err := client.Run(r.command)
			if err != nil {
				r.Lock()
				r.failures[host] = formatOutput(string(out))
				r.Unlock()
				return
			}
			r.Lock()
			r.successes[host] = formatOutput(string(out))
			r.Unlock()
		})
	}
	pool.StopAndWait()

	return nil
}

// PrintResult prints the results of the ssh command run
func (r *roll) printResult(failures bool) {

	output := color.New(color.FgWhite).SprintFunc()

	if !failures {
		for host, msg := range r.successes {
			fmt.Printf("  %s:\n%s\n", green(host), output(msg))
		}
	}

	for host, msg := range r.failures {
		fmt.Printf("  %s:\n%s\n", yellow(host), output(msg))
	}

	for host, msg := range r.errors {
		fmt.Printf("  %s:\n    %s\n\n", red(host), output(msg.Error()))
	}
	fmt.Printf("%s: %d %s: %d %s: %d\n", green("Success"), len(r.successes), yellow("Failure"), len(r.failures), red("Error"), len(r.errors))
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
