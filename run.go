package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

func run(c *cli.Context) error {
	if !c.Args().Present() {
		// cli.ShowCommandHelpAndExit(c, "run", 1)
		return fmt.Errorf("you need to provide a command to run")
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

	log.Info("Fetching list of GCE instances")
	instances, err := client.GetInstances(filter)
	if err != nil {
		return err
	}
	if len(instances) == 0 {
		log.Warn("No instances found")
		return nil
	}

	log.Info(fmt.Sprintf("Running [%s]", color.BlueString(cmd)))
	timeout, err := time.ParseDuration("10s")
	if err != nil {
		return err
	}

	batch := newRoll(cmd, timeout)
	err = batch.execute(instances, privateKeyPath, ignoreFingerprint)
	if err != nil {
		return err
	}

	batch.printResult(onlyFailures)
	return nil
}
