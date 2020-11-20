package gcp

import (
	"context"
	"fmt"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/networkmanagement/v1"
)

var computeService *compute.Service
var vpcService *networkmanagement.Service

// ComputeInit will initialize a GCP compute API client
func ComputeInit() error {
	var err error
	ctx := context.Background()

	computeService, err = compute.NewService(ctx)
	if err != nil {
		err = fmt.Errorf("Couldn't initialize GCP client (Compute): %v", err)
		return err
	}
	return nil
}

// VpcInit will initialize a GCP networkmanagement API client
func VpcInit() error {
	var err error
	ctx := context.Background()

	vpcService, err = networkmanagement.NewService(ctx)
	if err != nil {
		err = fmt.Errorf("Couldn't initialize GCP client (VPC): %v", err)
		return err
	}
	return nil
}
