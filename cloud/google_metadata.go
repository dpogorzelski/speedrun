package gcp

import (
	"fmt"
	"os"
	"os/user"
	"sort"
	"strings"

	"golang.org/x/crypto/ssh"
	"google.golang.org/api/compute/v1"
)

// AddKeyToMetadataP updates SSH key entires in the project metadata
func (c *ComputeClient) AddKeyToMetadataP(pubKey ssh.PublicKey) error {
	getProject := c.Projects.Get(c.Project)
	projectData, err := getProject.Do()
	if err != nil {
		return err
	}

	authorizedKey, err := formatPubKey(pubKey)
	if err != nil {
		return err
	}

	item, err := createMetadataItem(authorizedKey)
	if err != nil {
		return err
	}

	hasKey, same, i := hasItem(projectData.CommonInstanceMetadata, item)
	var items []*compute.MetadataItems

	if hasKey && same {
		return nil
	} else if hasKey && !same {
		items = updateMetadata(projectData.CommonInstanceMetadata, item, i)
	} else if !hasKey {
		items = appendToMetadata(projectData.CommonInstanceMetadata, item)
	}

	metadata := compute.Metadata{
		Fingerprint: projectData.CommonInstanceMetadata.Fingerprint,
		Items:       items,
	}

	setMetadata := c.Projects.SetCommonInstanceMetadata(c.Project, &metadata)
	_, err = setMetadata.Do()
	if err != nil {
		return err
	}

	return nil
}

// AddKeyToMetadata adds ssh public key to the intsance metadata
func (c *ComputeClient) AddKeyToMetadata(instance *compute.Instance, pubKey ssh.PublicKey) error {
	authorizedKey, err := formatPubKey(pubKey)
	if err != nil {
		return err
	}

	item, err := createMetadataItem(authorizedKey)
	if err != nil {
		return err
	}

	hasKey, same, i := hasItem(instance.Metadata, item)
	var items []*compute.MetadataItems

	if hasKey && same {
		return nil
	} else if hasKey && !same {
		items = updateMetadata(instance.Metadata, item, i)
	} else if !hasKey {
		items = appendToMetadata(instance.Metadata, item)
	}

	metadata := compute.Metadata{
		Fingerprint: instance.Metadata.Fingerprint,
		Items:       items,
	}

	instance.Metadata = &metadata
	s := strings.Split(instance.Zone, "/")
	zone := s[len(s)-1]
	call := c.Instances.Update(c.Project, zone, instance.Name, instance)
	_, err = call.Do()
	if err != nil {
		return fmt.Errorf("%s failed to update metadata: ", err)
	}
	return nil
}

func isBlocking(i *compute.Instance) bool {
	for _, m := range i.Metadata.Items {
		if m.Key == "sshKeys" || m.Key == "block-project-ssh-keys" {
			return true
		}
	}
	return false
}

func formatPubKey(pubKey ssh.PublicKey) (string, error) {
	authorizedKey := ssh.MarshalAuthorizedKey(pubKey)
	tk := strings.TrimSuffix(string(authorizedKey), "\n")
	return tk, nil
}

// Extracts username, algorithm and comment from a metadata SSH key item
func parseMetadataitem(key string) (string, string, string) {
	t := strings.Split(key, " ")
	head, comment := t[0], t[len(t)-1]
	username := strings.Split(head, ":")[0]
	algo := strings.Split(head, ":")[1]
	return username, algo, comment
}

// Verifies if a metadata item already exists for a given user/cipher/comment combination.
// If true it also returns the index number at which the existing item can be found otherwise index is -1.
func hasItem(md *compute.Metadata, x string) (bool, bool, int) {
	flatMD := flattenMetadata(md)
	if flatMD["ssh-keys"] == nil {
		return false, false, -1
	}

	items := strings.Split(flatMD["ssh-keys"].(string), "\n")
	username, algo, comment := parseMetadataitem(x)

	for i, e := range items {
		header := fmt.Sprintf("%s:%s", username, algo)
		if x == e {
			return true, true, i
		} else if strings.HasPrefix(e, header) && strings.HasSuffix(e, comment) {
			return true, false, i
		}
	}
	return false, false, -1
}

// createMetadataItem formats public key item according to GCP guidelines
func createMetadataItem(pubKey string) (string, error) {
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

func appendToMetadata(md *compute.Metadata, item string) []*compute.MetadataItems {
	var items []string
	flatMD := flattenMetadata(md)
	if flatMD["ssh-keys"] == nil {
		items = append(items, item)
		flatMD["ssh-keys"] = strings.Join(items, "\n")
		return expandComputeMetadata(flatMD)
	}

	items = strings.Split(flatMD["ssh-keys"].(string), "\n")
	items = append(items, item)
	flatMD["ssh-keys"] = strings.Join(items, "\n")
	return expandComputeMetadata(flatMD)
}

func updateMetadata(md *compute.Metadata, item string, i int) []*compute.MetadataItems {
	var items []string
	flatMD := flattenMetadata(md)
	items = strings.Split(flatMD["ssh-keys"].(string), "\n")
	items[i] = item
	flatMD["ssh-keys"] = strings.Join(items, "\n")
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
