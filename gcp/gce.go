package gcp

import (
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/compute/v1"
)

// GetIPAddresses returns a list of external IP addresses used for the SHH connection
func GetIPAddresses(instances []*compute.Instance) []string {
	log.Info("Fetching list of external IP addresses")
	addresses := []string{}
	for _, instance := range instances {
		addresses = append(addresses, instance.NetworkInterfaces[0].AccessConfigs[0].NatIP+":22")
	}

	return addresses
}

// GetInstances returns a list of external IP addresses used for the SHH connection
func GetInstances(project string, filter string) ([]*compute.Instance, error) {
	listCall := computeService.Instances.AggregatedList(project)
	listCall.Filter(filter)
	list, err := listCall.Do()
	if err != nil {
		return nil, err
	}

	instances := []*compute.Instance{}
	for _, item := range list.Items {
		for _, instance := range item.Instances {
			instances = append(instances, instance)
		}
	}

	return instances, nil
}
