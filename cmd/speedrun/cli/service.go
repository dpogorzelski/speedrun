package cli

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/alitto/pond"
	"github.com/apex/log"
	itls "github.com/speedrunsh/speedrun/pkg/common/tls"
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
	PreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
	RunE: action,
}

var startCmd = &cobra.Command{
	Use:     "start <servicename>",
	Short:   "Start a service",
	Example: "  speedrun service start nginx",
	Args:    cobra.MinimumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
	RunE: action,
}

var stopCmd = &cobra.Command{
	Use:     "stop <servicename>",
	Short:   "Stop a service",
	Example: "  speedrun service stop nginx",
	Args:    cobra.MinimumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
	RunE: action,
}

var statusCmd = &cobra.Command{
	Use:     "status <servicename>",
	Short:   "Return the status of the service",
	Example: "  speedrun service status nginx",
	Args:    cobra.MinimumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
	RunE: action,
}

func init() {
	serviceCmd.SetUsageTemplate(usage)
	serviceCmd.AddCommand(restartCmd)
	serviceCmd.AddCommand(startCmd)
	serviceCmd.AddCommand(stopCmd)
	serviceCmd.AddCommand(statusCmd)
	serviceCmd.PersistentFlags().StringP("target", "t", "", "Select instances that match the given criteria")
	serviceCmd.PersistentFlags().String("projectid", "", "Override GCP project id")
	serviceCmd.PersistentFlags().Bool("insecure", true, "Skip Portal's certificate verification (gRPC/QUIC)")
	serviceCmd.PersistentFlags().Bool("use-private-ip", false, "Connect to private IPs instead of public ones")
	viper.BindPFlag("transport.insecure", serviceCmd.PersistentFlags().Lookup("insecure"))
	viper.BindPFlag("portal.use-private-ip", serviceCmd.PersistentFlags().Lookup("use-private-ip"))
	viper.BindPFlag("gcp.projectid", serviceCmd.PersistentFlags().Lookup("projectid"))
}

func action(cmd *cobra.Command, args []string) error {
	project := viper.GetString("gcp.projectid")
	// insecure := viper.GetBool("transport.insecure")
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

	tlsConfig, err := itls.GenerateTLSConfig()
	if err != nil {
		return err
	}

	pool := pond.New(1000, 10000)
	for _, i := range instances {
		instance := i
		pool.Submit(func() {
			fields := log.Fields{
				"host":    instance.Name,
				"address": instance.GetAddress(usePrivateIP),
			}

			addr := fmt.Sprintf("%s:%d", instance.GetAddress(usePrivateIP), 1337)
			rawconn, err := tls.Dial("tcp", addr, tlsConfig)
			if err != nil {
				log.WithFields(fields).Error(err.Error())
				return
			}

			conn := drpcconn.New(rawconn)
			defer conn.Close()

			c := portalpb.NewDRPCPortalClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()

			var r *portalpb.Response
			switch cmd.Name() {
			case "restart":
				r, err = c.ServiceRestart(ctx, &portalpb.Service{Name: strings.Join(args, " ")})
			case "start":
				r, err = c.ServiceStart(ctx, &portalpb.Service{Name: strings.Join(args, " ")})
			case "stop":
				r, err = c.ServiceStop(ctx, &portalpb.Service{Name: strings.Join(args, " ")})
			case "status":
				r, err = c.ServiceStatus(ctx, &portalpb.Service{Name: strings.Join(args, " ")})
			}
			if err != nil {
				log.WithFields(fields).Warn(err.Error())
				return
			}
			log.WithFields(fields).Info(r.GetContent())
		})
	}
	pool.StopAndWait()
	return nil
}
