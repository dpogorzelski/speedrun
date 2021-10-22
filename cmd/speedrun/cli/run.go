package cli

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/alitto/pond"
	"github.com/speedrunsh/speedrun/pkg/common/cryptoutil"
	"github.com/speedrunsh/speedrun/pkg/speedrun/cloud"
	portalpb "github.com/speedrunsh/speedrun/proto/portal"
	"storj.io/drpc/drpcconn"

	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var runCmd = &cobra.Command{
	Use:     "run <command to run>",
	Short:   "Run a shell command on remote servers",
	Example: "  speedrun run whoami\n  speedrun run whoami --only-failures --target \"labels.foo = bar AND labels.environment = staging\"",
	Args:    cobra.MinimumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
	RunE: run,
}

func init() {
	runCmd.SetUsageTemplate(usage)
	runCmd.Flags().StringP("target", "t", "", "Fetch instances that match the target selection criteria")
	runCmd.Flags().String("projectid", "", "Override GCP project id")
	runCmd.Flags().Bool("insecure", false, "Skip server certificate verification")
	runCmd.Flags().String("ca", "ca.crt", "Path to the CA cert")
	runCmd.Flags().String("cert", "cert.crt", "Path to the client cert")
	runCmd.Flags().String("key", "key.key", "Path to the client key")
	runCmd.Flags().Bool("use-private-ip", false, "Connect to private IPs instead of public ones")

	viper.BindPFlag("gcp.projectid", runCmd.Flags().Lookup("projectid"))
	viper.BindPFlag("tls.insecure", runCmd.Flags().Lookup("insecure"))
	viper.BindPFlag("tls.ca", runCmd.Flags().Lookup("ca"))
	viper.BindPFlag("tls.cert", runCmd.Flags().Lookup("cert"))
	viper.BindPFlag("tls.key", runCmd.Flags().Lookup("key"))
	viper.BindPFlag("portal.use-private-ip", runCmd.Flags().Lookup("use-private-ip"))
}

func run(cmd *cobra.Command, args []string) error {
	command := strings.Join(args, " ")
	s := strings.Split(command, " ")
	project := viper.GetString("gcp.projectid")
	insecure := viper.GetBool("tls.insecure")
	caPath := viper.GetString("tls.ca")
	certPath := viper.GetString("tls.cert")
	keyPath := viper.GetString("tls.key")
	usePrivateIP := viper.GetBool("portal.use-private-ip")
	target, err := cmd.Flags().GetString("target")
	if err != nil {
		return err
	}

	gcpClient, err := cloud.NewGCPClient()
	if err != nil {
		return err
	}

	log.Info("Fetching instance list")
	instances, err := gcpClient.GetInstances(project, target)
	if err != nil {
		return err
	}

	if len(instances) == 0 {
		log.Warn("No instances found")
		return nil
	}

	var tlsConfig *tls.Config
	if insecure {
		tlsConfig, err = cryptoutil.InsecureTLSConfig()
		if err != nil {
			log.Fatalf("failed to generate tls config: %v", err)
		}
		log.Warn("Using insecure TLS configuration, this should be avoided in production environments")
	} else {
		tlsConfig, err = cryptoutil.ClientTLSConfig(caPath, certPath, keyPath)
		if err != nil {
			log.Fatalf("failed to generate tls config: %v", err)
		}
	}

	pool := pond.New(1000, 10000)
	for _, i := range instances {
		instance := i
		pool.Submit(func() {
			fields := log.Fields{
				"host":    instance.Name,
				"address": instance.GetAddress(usePrivateIP),
			}
			log := log.WithFields(fields)

			addr := fmt.Sprintf("%s:%d", instance.GetAddress(usePrivateIP), 1337)
			rawconn, err := tls.Dial("tcp", addr, tlsConfig)
			if err != nil {
				log.Error(err.Error())
				return
			}

			conn := drpcconn.New(rawconn)
			defer conn.Close()

			c := portalpb.NewDRPCPortalClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()

			r, err := c.RunCommand(ctx, &portalpb.CommandRequest{Name: s[0], Args: s[1:]})
			if err != nil {
				log.Error(err.Error())
				return
			}
			log.WithField("state", r.GetState()).Info(r.GetMessage())
		})
	}
	pool.StopAndWait()
	return nil
}
