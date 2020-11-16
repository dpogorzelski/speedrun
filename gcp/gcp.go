package gcp

import (
	"context"
	"speedrun/helpers"

	"google.golang.org/api/compute/v1"
)

var computeService *compute.Service

func init() {
	var err error

	ctx := context.Background()
	computeService, err = compute.NewService(ctx)
	if err != nil {
		helpers.Error(err.Error())
	}
}
