package gcp

import (
	"context"

	"google.golang.org/api/compute/v1"
)

// GetIPAddresses returns a list of external IP addresses used for the SHH connection
func (c *ComputeClient) GetIPAddresses(instances []*compute.Instance) []string {
	addresses := []string{}
	for _, instance := range instances {
		addresses = append(addresses, instance.NetworkInterfaces[0].AccessConfigs[0].NatIP+":22")
	}
	return addresses
}

// GetInstances returns a list of external IP addresses used for the SHH connection
func (c *ComputeClient) GetInstances(filter string) ([]*compute.Instance, error) {
	listCall := c.Instances.AggregatedList(c.Project)
	var ctx context.Context
	instances := []*compute.Instance{}

	listCall.Filter(filter).Pages(ctx, func(list *compute.InstanceAggregatedList) error {
		for _, item := range list.Items {
			instances = append(instances, item.Instances...)
		}
		return nil
	})
	_, err := listCall.Do()
	if err != nil {
		return nil, err
	}

	return instances, nil
}
