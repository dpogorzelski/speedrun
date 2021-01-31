package main

import (
	"context"
	"fmt"

	"google.golang.org/api/compute/v1"
)

// ComputeClient wraps the original *compute.Service
type ComputeClient struct {
	*compute.Service
	Project string
}

// NewComputeClient will initialize a GCP compute API client
func NewComputeClient(project string) (*ComputeClient, error) {
	var err error
	ctx := context.Background()

	s, err := compute.NewService(ctx)
	if err != nil {
		err = fmt.Errorf("Couldn't initialize GCP client (Compute): %v", err)
		return nil, err
	}
	computeService := &ComputeClient{s, project}

	return computeService, nil
}
