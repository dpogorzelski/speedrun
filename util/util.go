package util

import (
	"path/filepath"

	"github.com/apex/log"
	"github.com/mitchellh/go-homedir"
)

// DetermineKeyFilePath returns full path to the private key or an error otherwise
func DetermineKeyFilePath() (string, error) {
	log.Debug("Determining private key path")
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(home, ".speedrun/privatekey")
	return path, nil
}
