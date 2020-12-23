package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/fatih/color"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

func run(c *cli.Context) error {
	if c.Args().Len() < 1 {
		return cli.Exit("missing required command arguments", 1)
	}
	project := viper.GetString("gcp.projectid")

	client, err := NewComputeClient(project)
	if err != nil {
		return cli.Exit(err, 1)
	}

	filter := c.String("filter")
	onlyFailures := c.Bool("only-failures")

	pubKey, privKey, err := loadKeyPair()
	if err != nil {
		return cli.Exit(err, 1)
	}

	if err = client.getFirewallRules(); err != nil {
		return cli.Exit(err, 1)
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

	p.Start("Updating project metadata")
	err = client.UpdateProjectMetadata(pubKey)
	if err != nil {
		p.Error(err)
	}
	p.Stop()

	p.Start("Updating instance metadata")
	batch := 50
	for i := 0; i < len(instances); i += batch {
		j := i + batch
		if j > len(instances) {
			j = len(instances)
		}
		var wg sync.WaitGroup
		for a := range instances[i:j] {
			wg.Add(1)
			go client.UpdateInstanceMetadata(&wg, instances[a+i], pubKey)
		}
		wg.Wait()
	}
	p.Stop()

	p.Start(fmt.Sprintf("Running [%s]", color.BlueString(c.Args().First())))
	roll := newRoll(c.Args().First())
	err = roll.execute(instances, privKey)
	if err != nil {
		p.Error(err)
		os.Exit(1)
	}
	p.Stop()
	roll.printResult(onlyFailures)
	return nil
}
