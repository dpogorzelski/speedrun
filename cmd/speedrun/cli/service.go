package cli

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/alitto/pond"
	"github.com/apex/log"
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
}

func action(cmd *cobra.Command, args []string) error {
	usePrivateIP := viper.GetBool("portal.use-private-ip")

	tlsConfig, err := cloud.SetupTLS()
	if err != nil {
		return err
	}

	target, err := cmd.Flags().GetString("target")
	if err != nil {
		return err
	}

	portals, err := cloud.GetInstances(target)
	if err != nil {
		return err
	}

	pool := pond.New(1000, 10000)
	for _, p := range portals {
		portal := p
		pool.Submit(func() {
			fields := log.Fields{
				"host":    portal.Name,
				"address": portal.GetAddress(usePrivateIP),
			}
			log := log.WithFields(fields)

			addr := fmt.Sprintf("%s:%d", portal.GetAddress(usePrivateIP), 1337)
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
