package main

import (
	"os"

	"github.com/apex/log"
	loghandler "github.com/apex/log/handlers/text"
	"github.com/speedrunsh/speedrun/cmd/portal/cli"
)

func main() {
	h := loghandler.New(os.Stdout)
	log.SetHandler(h)
	cli.Execute()
}
