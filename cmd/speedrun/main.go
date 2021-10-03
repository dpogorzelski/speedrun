package main

import (
	"os"

	"github.com/speedrunsh/speedrun/cmd/speedrun/cli"

	"github.com/apex/log"
	loghandler "github.com/apex/log/handlers/text"
)

func main() {
	h := loghandler.New(os.Stdout)
	log.SetHandler(h)
	cli.Execute()
}
