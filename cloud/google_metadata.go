package cloud

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/user"
	"sort"
	"strings"

	"github.com/apex/log"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/oslogin/v1"
)

func (c *gcpClient) addUserKey(authorizedKey []byte) error {
	parent := fmt.Sprintf("users/%s", c.client_email)

	key := strings.TrimSuffix(string(authorizedKey), "\n")
	sshPublicKey := &oslogin.SshPublicKey{
		Key: string(key),
	}

	_, err := c.oslogin.Users.ImportSshPublicKey(parent, sshPublicKey).Do()
	return err
}

func (c *gcpClient) removeUserKey(authorizedKey []byte) error {
	key := strings.TrimSuffix(string(authorizedKey), "\n")
	name := fmt.Sprintf("users/%s/sshPublicKeys/%x", c.client_email, sha256.Sum256([]byte(key)))

	_, err := c.oslogin.Users.SshPublicKeys.Delete(name).Do()
	return err
}

func (c *gcpClient) listUserKeys() error {
	parent := fmt.Sprintf("users/%s", c.client_email)

	profile, err := c.oslogin.Users.GetLoginProfile(parent).Do()
	if err != nil {
		return err
	}

	for _, k := range profile.SshPublicKeys {
		log.Info(k.Key)
	}
	return nil
}

// AddKeyToMetadataP updates SSH key entires in the project metadata
func (c *gcpClient) addKeyToMetadata(authorizedKey []byte) error {
	getProject := c.gce.Projects.Get(c.Project)
	projectData, err := getProject.Do()
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

	setMetadata := c.gce.Projects.SetCommonInstanceMetadata(c.Project, &metadata)
	_, err = setMetadata.Do()
	if err != nil {
		return err
	}

	return nil
}

// removeKeyFromMetadata removes user's ssh public key from the project metadata
func (c *gcpClient) removeKeyFromMetadata(authorizedKey []byte) error {
	getProject := c.gce.Projects.Get(c.Project)
	projectData, err := getProject.Do()
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
		removeFromMetadata(projectData.CommonInstanceMetadata, item, i)
	} else {
		return nil
	}

	metadata := compute.Metadata{
		Fingerprint: projectData.CommonInstanceMetadata.Fingerprint,
		Items:       items,
	}

	setMetadata := c.gce.Projects.SetCommonInstanceMetadata(c.Project, &metadata)
	_, err = setMetadata.Do()
	if err != nil {
		return err
	}

	return nil
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
func createMetadataItem(authorizedKey []byte) (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}

	trimmedKey := strings.TrimSuffix(string(authorizedKey), "\n")

	v := fmt.Sprintf("%s:%s %s", user.Username, trimmedKey, hostname)
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

func removeFromMetadata(md *compute.Metadata, item string, i int) []*compute.MetadataItems {
	var items []string
	flatMD := flattenMetadata(md)
	items = strings.Split(flatMD["ssh-keys"].(string), "\n")
	copy(items[i:], items[i+1:])
	items[len(items)-1] = ""
	items = items[:len(items)-1]
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
