package key

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/pem"
	"io/ioutil"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/melbahja/goph"
	"github.com/mikesmitty/edkey"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"
)

const Comment = "speedrun"

type Key struct {
	User    string
	Comment string
	Key     []byte
}

func New() (*Key, error) {
	key := &Key{}
	_, rawKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	pemKey := &pem.Block{
		Type:  "OPENSSH PRIVATE KEY",
		Bytes: edkey.MarshalED25519PrivateKey(rawKey),
	}
	privateKey := pem.EncodeToMemory(pemKey)
	key.Key = privateKey

	user, err := user.Current()
	if err != nil {
		return nil, err
	}
	key.User = user.Username
	key.Comment = Comment
	return key, nil
}

func Read(path string) (*Key, error) {
	cleanPath := filepath.Clean(path)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return nil, err
	}

	file, err := ioutil.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	b64Result, err := base64.StdEncoding.DecodeString(string(file))
	if err != nil {
		return nil, err
	}

	key := &Key{}
	buf := bytes.NewBuffer(b64Result)
	enc := gob.NewDecoder(buf)

	err = enc.Decode(key)
	if err != nil {
		return nil, err
	}

	return key, nil

}

// Path returns full path to the private key or an error otherwise
func Path() (string, error) {
	log.Debug("Determining private key path")
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(home, ".speedrun/privatekey")
	return path, nil
}

func (k *Key) Write(path string) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(k)
	if err != nil {
		return err
	}

	b64Result := base64.StdEncoding.EncodeToString(buf.Bytes())
	if err != nil {
		return err
	}

	log.Debugf("Writing priviate key to %s", path)
	err = ioutil.WriteFile(path, []byte(b64Result), 0600)
	if err != nil {
		return err
	}
	return nil

}

func (k *Key) MarshalAuthorizedKey() (string, error) {
	privKey, err := ssh.ParsePrivateKey(k.Key)
	if err != nil {
		return "", err
	}

	authorizedKey := ssh.MarshalAuthorizedKey(privKey.PublicKey())
	trimmedKey := strings.TrimSuffix(string(authorizedKey), "\n")
	return trimmedKey, nil
}

func (k *Key) GetAuth() (goph.Auth, error) {
	privKey, err := ssh.ParsePrivateKey(k.Key)
	if err != nil {
		return nil, err
	}

	return goph.Auth{
		ssh.PublicKeys(privKey),
	}, nil
}
