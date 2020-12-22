package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name: "speedrun",
		Commands: []*cli.Command{
			{
				Name:   "init",
				Usage:  "Initialize speedrun",
				Action: initialize,
			},
			{
				Name:  "run",
				Usage: "Runs a command on remote servers",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "filter", DefaultText: "gcloud resource filter"},
					&cli.BoolFlag{Name: "only-failures", DefaultText: "print only failures and errors"},
				},
				Before: configInitialized,
				Action: run,
			},
			{
				Name:  "key",
				Usage: "Manage ssh keys",
				Subcommands: []*cli.Command{
					{
						Name:   "new",
						Usage:  "create a new ssh key",
						Before: configInitialized,
						Action: newKey,
					},
					{
						Name:   "show",
						Usage:  "show current ssh key",
						Before: configInitialized,
						Action: showKey,
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
