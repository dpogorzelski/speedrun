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

// // UpdateSSHKeys returns a dictionary of instance names and their sshKeys metadata entry
// func UpdateSSHKeys(instance *compute.Instance, key string) {
// 	log.Info("Fetching sshKeys from instance metadata")

// 	// computeService.Instances.Update()
// }

// GetInstances returns a list of external IP addresses used for the SHH connection
func GetInstances(project string, filter string) ([]*compute.Instance, error) {
	log.Info("Fetching list of GCE instances")
	listCall := computeService.Instances.AggregatedList(project)
	listCall.Filter(filter)
	list, err := listCall.Do()
	if err != nil {
		return nil, err
	}

	instances := []*compute.Instance{}
	for _, item := range list.Items {
		for _, instance := range item.Instances {
			// for _, m := range instance.Metadata.Items {
			// 	if m.Key == "sshKeys" || m.Key == "block-project-ssh-keys" {
			// 		log.Debugln(instance.Name, "Ignoring, this instance is blocking project wide SSH keys")
			// 	}
			// }
			instances = append(instances, instance)
		}
	}

	return instances, nil
}
