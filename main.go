package main

import (
	"os"
	"speedrun/cmd"

	"github.com/apex/log"
	loghandler "github.com/apex/log/handlers/cli"
)

func main() {
	h := loghandler.New(os.Stdout)
	h.Padding = 0
	log.SetHandler(h)
	cmd.Execute()
}
