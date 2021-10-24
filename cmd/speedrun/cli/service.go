package cli

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/alitto/pond"
	"github.com/apex/log"
	"github.com/speedrunsh/speedrun/pkg/common/cryptoutil"
	"github.com/speedrunsh/speedrun/pkg/speedrun/cloud"
	portalpb "github.com/speedrunsh/speedrun/proto/portal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"storj.io/drpc/drpcconn"
)

var serviceCmd = &cobra.Command{
	Use:              "service",
	Short:            "Manage services",
	TraverseChildren: true,
}

var restartCmd = &cobra.Command{
	Use:     "restart <servicename>",
	Short:   "Restart a service",
	Example: "  speedrun service restart nginx",
	Args:    cobra.MinimumNArgs(1),
	RunE:    action,
}

var startCmd = &cobra.Command{
	Use:     "start <servicename>",
	Short:   "Start a service",
	Example: "  speedrun service start nginx",
	Args:    cobra.MinimumNArgs(1),
	RunE:    action,
}

var stopCmd = &cobra.Command{
	Use:     "stop <servicename>",
	Short:   "Stop a service",
	Example: "  speedrun service stop nginx",
	Args:    cobra.MinimumNArgs(1),
	RunE:    action,
}

var statusCmd = &cobra.Command{
	Use:     "status <servicename>",
	Short:   "Return the status of the service",
	Example: "  speedrun service status nginx",
	Args:    cobra.MinimumNArgs(1),
	RunE:    action,
}

func init() {
	serviceCmd.SetUsageTemplate(usage)
	serviceCmd.AddCommand(restartCmd)
	serviceCmd.AddCommand(startCmd)
	serviceCmd.AddCommand(stopCmd)
	serviceCmd.AddCommand(statusCmd)
	serviceCmd.PersistentFlags().StringP("target", "t", "", "Select instances that match the given criteria")
	serviceCmd.PersistentFlags().String("projectid", "", "Override GCP project id")
	serviceCmd.PersistentFlags().Bool("insecure", false, "Skip server certificate verification")
	serviceCmd.PersistentFlags().String("ca", "ca.crt", "Path to the CA cert")
	serviceCmd.PersistentFlags().String("cert", "cert.crt", "Path to the client cert")
	serviceCmd.PersistentFlags().String("key", "key.key", "Path to the client key")
	serviceCmd.PersistentFlags().Bool("use-private-ip", false, "Connect to private IPs instead of public ones")

	viper.BindPFlag("gcp.projectid", serviceCmd.PersistentFlags().Lookup("projectid"))
	viper.BindPFlag("tls.insecure", serviceCmd.PersistentFlags().Lookup("insecure"))
	viper.BindPFlag("tls.ca", serviceCmd.PersistentFlags().Lookup("ca"))
	viper.BindPFlag("tls.cert", serviceCmd.PersistentFlags().Lookup("cert"))
	viper.BindPFlag("tls.key", serviceCmd.PersistentFlags().Lookup("key"))
	viper.BindPFlag("portal.use-private-ip", serviceCmd.PersistentFlags().Lookup("use-private-ip"))
}

func action(cmd *cobra.Command, args []string) error {
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
		log.Warn("Using insecure TLS configuration, this should be avoided in production environments")
		tlsConfig, err = cryptoutil.InsecureTLSConfig()
	} else {
		tlsConfig, err = cryptoutil.ClientTLSConfig(caPath, certPath, keyPath)
	}
	if err != nil {
		return fmt.Errorf("could not initialize TLS config: %v", err)
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

			switch cmd.Name() {
			case "restart":
				r, err := c.ServiceRestart(ctx, &portalpb.ServiceRequest{Name: strings.Join(args, " ")})
				if err != nil {
					log.Error(err.Error())
					return
				}
				log.WithField("state", r.GetState()).Info(r.GetMessage())
			case "start":
				r, err := c.ServiceStart(ctx, &portalpb.ServiceRequest{Name: strings.Join(args, " ")})
				if err != nil {
					log.Error(err.Error())
					return
				}
				log.WithField("state", r.GetState()).Info(r.GetMessage())
			case "stop":
				r, err := c.ServiceStop(ctx, &portalpb.ServiceRequest{Name: strings.Join(args, " ")})
				if err != nil {
					log.Error(err.Error())
					return
				}
				log.WithField("state", r.GetState()).Info(r.GetMessage())
			case "status":
				r, err := c.ServiceStatus(ctx, &portalpb.ServiceRequest{Name: strings.Join(args, " ")})
				if err != nil {
					log.Error(err.Error())
					return
				}
				log.WithField("state", r.GetState()).Infof("Service status is: loadstate: %s activestate: %s substate: %s", r.GetLoadstate(), r.GetActivestate(), r.GetSubstate())
			}

		})
	}
	pool.StopAndWait()
	return nil
}
