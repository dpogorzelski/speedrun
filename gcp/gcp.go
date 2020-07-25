package gcp

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/api/compute/v1"
)

var computeService *compute.Service

func init() {
	var err error

	ctx := context.Background()
	computeService, err = compute.NewService(ctx)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
