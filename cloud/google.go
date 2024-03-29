package cloud

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/oslogin/v1"
)

type GCPClient struct {
	gce          *compute.Service
	oslogin      *oslogin.Service
	client_email string
	Project      string
}

func NewGCPClient(project string) (*GCPClient, error) {
	var err error
	ctx := context.Background()
	credentials, err := google.FindDefaultCredentials(ctx, compute.ComputeScope)
	if err != nil {
		err = fmt.Errorf("couldn't fetch default client credentials: %v", err)
		return nil, err
	}

	var jsonCreds map[string]interface{}
	err = json.Unmarshal(credentials.JSON, &jsonCreds)
	if err != nil {
		err = fmt.Errorf("couldn't decode default client credentials json: %v", err)
		return nil, err
	}

	gce, err := compute.NewService(ctx)
	if err != nil {
		err = fmt.Errorf("couldn't initialize GCP client: %v", err)
		return nil, err
	}

	osc, err := oslogin.NewService(ctx)
	if err != nil {
		return nil, err
	}

	c := &GCPClient{
		gce:          gce,
		oslogin:      osc,
		client_email: jsonCreds["client_email"].(string),
		Project:      project,
	}

	return c, nil
}
