package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
)

func main() {
	init := &cli.Command{
		Name:   "init",
		Usage:  "Initialize speedrun",
		Action: initialize,
	}

	run := &cli.Command{
		Name:  "run",
		Usage: "Runs a command on remote servers",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "filter", Usage: "gcloud resource filter"},
			altsrc.NewBoolFlag(&cli.BoolFlag{Name: "only-failures", Usage: "print only failures and errors"}),
			&cli.BoolFlag{Name: "private-ip", Usage: "connect to private IPs instead of public ones"},
			altsrc.NewBoolFlag(&cli.BoolFlag{Name: "ignore-fingerprint", Usage: "ignore host's fingerprint mismatch"}),
		},
		Before: configInitialized,
		Action: run,
		UsageText: "speedrun run [command options] <command to send>\n\n" +
			"EXAMPLES:\n" +
			"   speedrun run \"uname -r\"\n" +
			"   speedrun run --only-failures --filter \"labels.foo = bar AND labels.environment = staging\" \"uname -r\"",
	}

	key := &cli.Command{
		Name:  "key",
		Usage: "Manage ssh keys",
		Subcommands: []*cli.Command{
			{
				Name:   "new",
				Usage:  "Create a new ssh key",
				Before: configInitialized,
				Action: createKey,
			},
			{
				Name:   "show",
				Usage:  "Show current ssh key",
				Before: configInitialized,
				Action: showKey,
			},
			{
				Name:  "set",
				Usage: "Set key in the project's metadata",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "filter", Usage: "gcloud resource filter"},
				},
				Before: configInitialized,
				Action: setKey,
			},
		},
	}

	appFlags := []cli.Flag{
		&cli.StringFlag{
			Name:  "config",
			Usage: "The location of the speedrun config file",
			Value: configPath(),
		},
	}

	app := &cli.App{
		Name: "speedrun",
		Authors: []*cli.Author{{
			Name:  "Dawid Pogorzelski",
			Email: "dawid@pogorzelski.dev",
		}},
		Before: altsrc.InitInputSourceWithContext(appFlags, altsrc.NewTomlSourceFromFlagFunc("config")),
		Flags:  appFlags,
		Commands: []*cli.Command{
			init,
			run,
			key,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
