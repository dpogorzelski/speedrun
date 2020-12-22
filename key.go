package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func newKey(ctx *cli.Context) error {
	_, _, err := GenerateKeyPair()
	if err != nil {
		return cli.Exit(err, 1)
	}
	fmt.Println("generated new ssh key")
	return nil
}

func showKey(ctx *cli.Context) error {
	fmt.Println("showing private key")
	return nil
}
