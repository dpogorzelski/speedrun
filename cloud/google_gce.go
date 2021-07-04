package cloud

import (
	"context"

	"google.golang.org/api/compute/v1"
)

// GetInstances returns a list of external IP addresses used for the SHH connection
func (c *GCPClient) GetInstances(filter string, usePrivateIP bool) ([]Instance, error) {
	instances := []Instance{}
	listCall := c.gce.Instances.AggregatedList(c.Project).Fields("nextPageToken", "items(Name,NetworkInterfaces)")
	var ctx context.Context

	listCall.Filter(filter).Pages(ctx, func(list *compute.InstanceAggregatedList) error {
		for _, item := range list.Items {
			for _, instance := range item.Instances {
				i := &Instance{
					Name: instance.Name,
				}
				if usePrivateIP {
					i.Address = instance.NetworkInterfaces[0].NetworkIP
				} else {
					i.Address = instance.NetworkInterfaces[0].AccessConfigs[0].NatIP
				}

				instances = append(instances, *i)
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
