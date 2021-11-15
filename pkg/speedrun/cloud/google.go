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

// GetInstances returns a list of external IP addresses used for the SHH connection
func (c *GoogleClient) GetInstances(project string) ([]Instance, error) {
	instances := []Instance{}
	listCall := c.Instances.AggregatedList(project).Fields("nextPageToken", "items(Name,NetworkInterfaces,Labels)")
	var ctx context.Context

	listCall.Pages(ctx, func(list *compute.InstanceAggregatedList) error {
		for _, item := range list.Items {
			for _, instance := range item.Instances {
				i := Instance{
					Name:           instance.Name,
					PrivateAddress: instance.NetworkInterfaces[0].NetworkIP,
					PublicAddress:  instance.NetworkInterfaces[0].AccessConfigs[0].NatIP,
					Labels:         instance.Labels,
				}
				instances = append(instances, i)
			}
		}
		return nil
	})
	_, err := listCall.Do()
	if err != nil {
		return nil, err
	}

	return instances, nil
}
