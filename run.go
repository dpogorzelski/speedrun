package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

func run(c *cli.Context) error {
	if !c.Args().Present() {
		// cli.ShowCommandHelpAndExit(c, "run", 1)
		return cli.Exit("you need to provide a command to run", 1)
	}
	cmd := strings.Join(c.Args().Slice(), " ")

	client, err := NewComputeClient(config.Gcp.Projectid)
	if err != nil {
		return cli.Exit(err, 1)
	}

	filter := c.String("filter")
	onlyFailures := c.Bool("only-failures")
	ignoreFingerprint := c.Bool("ignore-fingerprint")

	privateKeyPath, err := determineKeyFilePath()
	if err != nil {
		return err
	}

	p := NewProgress()
	p.Start("Fetching list of GCE instances")
	instances, err := client.GetInstances(filter)
	if err != nil {
		p.Error(err)
	}
	if len(instances) == 0 {
		p.Failure("no instances found")
	}
	p.Stop()

	p.Start(fmt.Sprintf("Running [%s]", color.BlueString(cmd)))
	timeout, err := time.ParseDuration("10s")
	if err != nil {
		return err
	}

	batch := newRoll(cmd, timeout)
	err = batch.execute(instances, privateKeyPath, ignoreFingerprint)
	if err != nil {
		p.Error(err)
		os.Exit(1)
	}
	p.Stop()

	batch.printResult(onlyFailures)
	return nil
}
