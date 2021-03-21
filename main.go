package main

import (
	"os"

	"speedrun/cmd"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
)

func main() {
	h := cli.New(os.Stdout)
	h.Padding = 0
	log.SetHandler(h)
	cmd.Execute()
}
