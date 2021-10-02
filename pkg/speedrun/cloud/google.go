package cloud

import (
	"context"
	"fmt"

	"google.golang.org/api/compute/v1"
)

type GoogleClient struct {
	*compute.Service
}

func NewGCPClient() (*GoogleClient, error) {
	var err error
	ctx := context.Background()

	gce, err := compute.NewService(ctx)
	if err != nil {
		err = fmt.Errorf("couldn't initialize GCP client: %v", err)
		return nil, err
	}

	return &GoogleClient{gce}, nil
}
