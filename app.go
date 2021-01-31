package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

var config *Config

func main() {
	var err error
	config, err = NewConfig()
	if err != nil {
		cli.Exit(err, 1)
	}

	init := &cli.Command{
		Name:   "init",
		Usage:  "Initialize speedrun",
		Action: config.Create,
	}

	run := &cli.Command{
		Name:  "run",
		Usage: "Runs a command on remote servers",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "filter", Usage: "gcloud resource filter"},
			&cli.BoolFlag{Name: "only-failures", Usage: "print only failures and errors"},
			&cli.BoolFlag{Name: "private-ip", Usage: "connect to private IPs instead of public ones"},
			&cli.BoolFlag{Name: "ignore-fingerprint", Usage: "ignore host's fingerprint mismatch"},
		},
		Before: config.Read,
		Action: run,
		UsageText: "speedrun run [command options] -- <command to run>\n\n" +
			"EXAMPLES:\n" +
			"   speedrun run -- uname -r\n" +
			"   speedrun run --only-failures --filter \"labels.foo = bar AND labels.environment = staging\" -- uname -r",
	}

	key := &cli.Command{
		Name:  "key",
		Usage: "Manage ssh keys",
		Subcommands: []*cli.Command{
			{
				Name:   "new",
				Usage:  "Create a new ssh key",
				Before: config.Read,
				Action: createKey,
			},
			{
				Name:   "show",
				Usage:  "Show current ssh key",
				Before: config.Read,
				Action: showKey,
			},
			{
				Name:  "set",
				Usage: "Set key in the project's metadata",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "filter", Usage: "gcloud resource filter"},
				},
				Before: config.Read,
				Action: setKey,
			},
		},
	}

	app := &cli.App{
		Name:      "speedrun",
		Usage:     "Cloud first command execution",
		UsageText: "speedrun command [subcommand]",
		Commands: []*cli.Command{
			init,
			run,
			key,
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
