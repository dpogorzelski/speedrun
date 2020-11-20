package gcp

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
	"google.golang.org/api/compute/v1"
)

type computeClient struct {
	*compute.Service
	project string
}

var computeService *computeClient

// ComputeInit will initialize a GCP compute API client
func ComputeInit() error {
	project := viper.GetString("gcp.projectid")
	var err error
	ctx := context.Background()

	s, err := compute.NewService(ctx)
	if err != nil {
		err = fmt.Errorf("Couldn't initialize GCP client (Compute): %v", err)
		return err
	}
	computeService = &computeClient{s, project}

	return nil
}
