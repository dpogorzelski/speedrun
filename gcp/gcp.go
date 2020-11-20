package gcp

import (
	"context"
	"fmt"
	"speedrun/utils"

	"google.golang.org/api/compute/v1"
)

var computeService *compute.Service

func init() {
	var err error

	ctx := context.Background()
	computeService, err = compute.NewService(ctx)
	if err != nil {
		err = fmt.Errorf("Couldn't initialize GCP client: %v", err)
		utils.Error(err.Error())
	}
}
