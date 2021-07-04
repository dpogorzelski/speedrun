package cloud

import (
	"crypto/sha256"
	"fmt"
	"speedrun/key"

	"github.com/apex/log"
	"google.golang.org/api/oslogin/v1"
)

func (c *GCPClient) AddUserKey(key *key.Key) error {
	parent := fmt.Sprintf("users/%s", c.client_email)

	authorizedKey, err := key.MarshalAuthorizedKey()
	if err != nil {
		return err
	}

	sshPublicKey := &oslogin.SshPublicKey{
		Key: string(authorizedKey),
	}

	_, err = c.oslogin.Users.ImportSshPublicKey(parent, sshPublicKey).Do()
	return err
}

func (c *GCPClient) RemoveUserKey(key *key.Key) error {
	authorizedKey, err := key.MarshalAuthorizedKey()
	if err != nil {
		return err
	}

	name := fmt.Sprintf("users/%s/sshPublicKeys/%x", c.client_email, sha256.Sum256([]byte(authorizedKey)))
	_, err = c.oslogin.Users.SshPublicKeys.Delete(name).Do()
	return err
}

func (c *GCPClient) ListUserKeys() error {
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

func (c *GCPClient) GetSAUsername() (string, error) {
	parent := fmt.Sprintf("users/%s", c.client_email)

	profile, err := c.oslogin.Users.GetLoginProfile(parent).Do()
	if err != nil {
		return "", err
	}

	return profile.PosixAccounts[0].Username, nil
}
