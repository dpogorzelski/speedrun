package gcp

import (
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/compute/v1"
)

// GetIPAddresses returns a list of external IP addresses used for the SHH connection
func GetIPAddresses(instances *compute.InstanceAggregatedList) []string {
	log.Info("Fetching list of external IP addresses")
	addresses := []string{}
	for _, v := range instances.Items {
		for _, instance := range v.Instances {
			addresses = append(addresses, instance.NetworkInterfaces[0].AccessConfigs[0].NatIP+":22")
		}
	}
	return addresses
}

// GetInstances returns a list of external IP addresses used for the SHH connection
func GetInstances(project string, filter string) (*compute.InstanceAggregatedList, error) {
	log.Info("Fetching list of GCE instances")
	listCall := computeService.Instances.AggregatedList(project)
	listCall.Filter(filter)
	instances, err := listCall.Do()
	if err != nil {
		return nil, err
	}

	return instances, nil
}
