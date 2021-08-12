package cloud

import (
	"context"
	"crypto/sha256"
	"fmt"
	"speedrun/key"

	"github.com/apex/log"
	common "google.golang.org/genproto/googleapis/cloud/oslogin/common"
	osloginpb "google.golang.org/genproto/googleapis/cloud/oslogin/v1"
)

func (c *GCPClient) AddUserKey(ctx context.Context, key *key.Key) error {
	parent := fmt.Sprintf("users/%s", c.client_email)

	authorizedKey, err := key.MarshalAuthorizedKey()
	if err != nil {
		return err
	}

	k := &common.SshPublicKey{
		Key: string(authorizedKey),
	}

	req := &osloginpb.ImportSshPublicKeyRequest{
		Parent:       parent,
		SshPublicKey: k,
	}

	_, err = c.oslogin.ImportSshPublicKey(ctx, req)
	return err
}

func (c *GCPClient) RemoveUserKey(ctx context.Context, key *key.Key) error {
	authorizedKey, err := key.MarshalAuthorizedKey()
	if err != nil {
		return err
	}

	name := fmt.Sprintf("users/%s/sshPublicKeys/%x", c.client_email, sha256.Sum256([]byte(authorizedKey)))
	req := &osloginpb.DeleteSshPublicKeyRequest{Name: name}
	return c.oslogin.DeleteSshPublicKey(ctx, req)
}

func (c *GCPClient) ListUserKeys(ctx context.Context) error {
	parent := fmt.Sprintf("users/%s", c.client_email)

	req := &osloginpb.GetLoginProfileRequest{Name: parent}
	profile, err := c.oslogin.GetLoginProfile(ctx, req)
	if err != nil {
		return err
	}

	for _, k := range profile.SshPublicKeys {
		log.Info(k.Key)
	}
	return nil
}

func (c *GCPClient) GetSAUsername(ctx context.Context) (string, error) {
	parent := fmt.Sprintf("users/%s", c.client_email)

	req := &osloginpb.GetLoginProfileRequest{Name: parent}
	profile, err := c.oslogin.GetLoginProfile(ctx, req)
	if err != nil {
		return "", err
	}

	return profile.PosixAccounts[0].Username, nil
}
