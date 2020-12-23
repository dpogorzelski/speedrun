package main

import (
	"context"
	"fmt"
	"os/user"
	"strings"
	"time"

	"github.com/fatih/color"
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

func NewRun() *Run {
	r := Run{}
	r.res.errors = make(map[string]error)
	r.res.failures = make(map[string]string)
	r.res.successes = make(map[string]string)
	return &r
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
		host := instanceDict[resp.ID()]
		if err := resp.Err(); err != nil {
			run.res.errors[host] = err
			continue
		}

		output, _, _ := resp.GetText(vs)
		if resp.ExitStatus() == 0 {
			run.res.successes[host] = formatOutput(output)
		} else {
			run.res.failures[host] = formatOutput(output)
		}

	}
	return run, nil
}

// PrintResult prints the results of the ssh command run
func (r Run) PrintResult(failures bool) {

	output := color.New(color.FgWhite).SprintFunc()

	if !failures {
		for host, msg := range r.res.successes {
			fmt.Printf("  %s:\n%s\n", green(host), output(msg))
		}
	}

	for host, msg := range r.res.failures {
		fmt.Printf("  %s:\n%s\n", yellow(host), output(msg))
	}

	for host, msg := range r.res.errors {
		fmt.Printf("  %s:\n    %s\n\n", red(host), output(msg.Error()))
	}
	fmt.Printf("%s: %d %s: %d %s: %d\n", green("Success"), len(r.res.successes), yellow("Failure"), len(r.res.failures), red("Error"), len(r.res.errors))
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
