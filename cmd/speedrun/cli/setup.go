package cli

import (
	"crypto/tls"
	"fmt"

	"github.com/apex/log"
	"github.com/speedrunsh/speedrun/pkg/common/cryptoutil"
	"github.com/speedrunsh/speedrun/pkg/speedrun/cloud"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func getPortals(cmd *cobra.Command) ([]cloud.Instance, error) {
	project := viper.GetString("gcp.projectid")
	_, err := cmd.Flags().GetString("target")
	if err != nil {
		return nil, err
	}

	gcpClient, err := cloud.NewGCPClient()
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

func setupTLS() (*tls.Config, error) {
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
