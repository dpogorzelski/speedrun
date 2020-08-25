package gcp

import (
	"fmt"
	"os"
	"os/user"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"google.golang.org/api/compute/v1"
)

// func updateInstanceMetadata(project string, instance *compute.Instance, pubKey []byte) error {
// 	metaItems := append(instance.Metadata.Items, &compute.MetadataItems{})
// 	metadata := compute.Metadata{
// 		Fingerprint: instance.Metadata.Fingerprint,
// 		Items:       metaItems,
// 	}
// 	computeService.Instances.SetMetadata(project, instance.Zone, instance.Name, &metadata)
// 	return nil
// }

// UpdateProjectMetadata adds public key to the project metadata to allow SSH connections to GCE instances with the corresponding private key
func UpdateProjectMetadata(project string, pubKey ssh.PublicKey) error {
	authorizedKey, err := formatSSHPubKey(pubKey)
	if err != nil {
		return err
	}

	entry, err := createMetadataEntry(authorizedKey)
	if err != nil {
		return err
	}

	getProject := computeService.Projects.Get(project)
	projectData, err := getProject.Do()
	if err != nil {
		return err
	}

	has, same, i := hasEntry(projectData.CommonInstanceMetadata, entry)
	var items []*compute.MetadataItems

	if has && same {
		log.Info("Public key already present in metadata")
		return nil
	} else if has && !same {
		items = updateMetadata(projectData.CommonInstanceMetadata, entry, i)
	} else if !has {
		items = appendToMetadata(projectData.CommonInstanceMetadata, entry)
	}

	metadata := compute.Metadata{
		Fingerprint: projectData.CommonInstanceMetadata.Fingerprint,
		Items:       items,
	}

	setMetadata := computeService.Projects.SetCommonInstanceMetadata(project, &metadata)
	_, err = setMetadata.Do()
	if err != nil {
		return err
	}

	return nil
}

// UpdateInstanceMetadata adds ssh public key to the intsance metadata
func UpdateInstanceMetadata(project string, instances []*compute.Instance, pubKey ssh.PublicKey) error {
	authorizedKey, err := formatSSHPubKey(pubKey)
	if err != nil {
		return err
	}

	entry, err := createMetadataEntry(authorizedKey)
	if err != nil {
		return err
	}

	for _, instance := range instances {
		has, same, i := hasEntry(instance.Metadata, entry)
		var items []*compute.MetadataItems

		if has && same {
			log.Info("Public key already present in metadata")
			return nil
		} else if has && !same {
			items = updateMetadata(instance.Metadata, entry, i)
		} else if !has {
			items = appendToMetadata(instance.Metadata, entry)
		}

		metadata := compute.Metadata{
			Fingerprint: instance.Metadata.Fingerprint,
			Items:       items,
		}

		instance.Metadata = &metadata
		s := strings.Split(instance.Zone, "/")
		zone := s[len(s)-1]
		call := computeService.Instances.Update(project, zone, instance.Name, instance)
		_, err = call.Do()
		if err != nil {
			return err
		}
	}

	return nil
}

func formatSSHPubKey(pubKey ssh.PublicKey) (string, error) {
	authorizedKey := ssh.MarshalAuthorizedKey(pubKey)
	tk := strings.TrimSuffix(string(authorizedKey), "\n")
	return tk, nil
}

// Extracts username, algorithm and comment from a metadata SSH key entry
func parseMetadataEntry(key string) (string, string, string) {
	t := strings.Split(key, " ")
	head, comment := t[0], t[len(t)-1]
	username := strings.Split(head, ":")[0]
	algo := strings.Split(head, ":")[1]
	return username, algo, comment
}

// Verifies if a metadata entry already exists for a given user/cipher/comment combination.
// If true it also returns the index number at which the existing entry can be found otherwise index is -1.
func hasEntry(md *compute.Metadata, x string) (bool, bool, int) {
	flatMD := flattenMetadata(md)
	if flatMD["ssh-keys"] == nil {
		return false, false, -1
	}

	entries := strings.Split(flatMD["ssh-keys"].(string), "\n")
	username, algo, comment := parseMetadataEntry(x)

	for i, e := range entries {
		header := fmt.Sprintf("%s:%s", username, algo)
		if x == e {
			return true, true, i
		} else if strings.HasPrefix(e, header) && strings.HasSuffix(e, comment) {
			return true, false, i
		}
	}
	return false, false, -1
}

// createMetadataEntry formats public key entry according to GCP guidelines
func createMetadataEntry(pubKey string) (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}

	v := fmt.Sprintf("%s:%s %s", user.Username, pubKey, hostname)
	return v, nil
}

func appendToMetadata(md *compute.Metadata, entry string) []*compute.MetadataItems {
	log.Info("Appending new public key to metadata")

	var entries []string
	flatMD := flattenMetadata(md)
	if flatMD["ssh-keys"] == nil {
		entries = append(entries, entry)
		flatMD["ssh-keys"] = strings.Join(entries, "\n")
		return expandComputeMetadata(flatMD)
	}

	entries = strings.Split(flatMD["ssh-keys"].(string), "\n")
	entries = append(entries, entry)
	flatMD["ssh-keys"] = strings.Join(entries, "\n")
	return expandComputeMetadata(flatMD)
}

func updateMetadata(md *compute.Metadata, entry string, i int) []*compute.MetadataItems {
	log.Info("Updating existing metadata entry with new public key")

	var entries []string
	flatMD := flattenMetadata(md)
	entries = strings.Split(flatMD["ssh-keys"].(string), "\n")
	entries[i] = entry
	flatMD["ssh-keys"] = strings.Join(entries, "\n")
	return expandComputeMetadata(flatMD)
}

func expandComputeMetadata(m map[string]interface{}) []*compute.MetadataItems {
	metadata := make([]*compute.MetadataItems, len(m))
	var keys []string
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		v := m[key].(string)
		metadata = append(metadata, &compute.MetadataItems{
			Key:   key,
			Value: &v,
		})
	}

	return metadata
}

func flattenMetadata(metadata *compute.Metadata) map[string]interface{} {
	metadataMap := make(map[string]interface{})
	for _, item := range metadata.Items {
		metadataMap[item.Key] = *item.Value
	}
	return metadataMap
}
