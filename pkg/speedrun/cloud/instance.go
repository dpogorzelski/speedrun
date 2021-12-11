package cloud

import (
	"crypto/tls"
	"fmt"

	"github.com/apex/log"
	"github.com/speedrunsh/speedrun/pkg/common/cryptoutil"
	"github.com/spf13/viper"
)

type Instance struct {
	PublicAddress  string
	PrivateAddress string
	Name           string
	Labels         map[string]string
}

func (i Instance) GetAddress(private bool) string {
	if private {
		return i.PrivateAddress
	}

	return i.PublicAddress
}

func GetInstances(target string) ([]Instance, error) {
	project := viper.GetString("gcp.projectid")

	gcpClient, err := NewGCPClient()
	if err != nil {
		return nil, err
	}

	log.Info("Fetching instance list")
	instances, err := gcpClient.GetInstances(project)
	if err != nil {
		return nil, err
	}

	if len(instances) == 0 {
		return nil, fmt.Errorf("no instances found")
	}

	return instances, nil
}

func SetupTLS() (*tls.Config, error) {
	insecure := viper.GetBool("tls.insecure")
	caPath := viper.GetString("tls.ca")
	certPath := viper.GetString("tls.cert")
	keyPath := viper.GetString("tls.key")

	if insecure {
		log.Warn("Using insecure TLS configuration, this should be avoided in production environments")
		return cryptoutil.InsecureTLSConfig()
	} else {
		return cryptoutil.ClientTLSConfig(caPath, certPath, keyPath)
	}

}
