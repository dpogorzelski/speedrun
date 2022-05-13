package cli

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/alitto/pond"
	"github.com/apex/log"
	"github.com/dpogorzelski/speedrun/pkg/speedrun/cloud"
	portalpb "github.com/dpogorzelski/speedrun/proto/portal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"storj.io/drpc/drpcconn"
)

var systemCmd = &cobra.Command{
	Use:              "system",
	Short:            "Manage the system",
	TraverseChildren: true,
}

var rebootCmd = &cobra.Command{
	Use:     "reboot",
	Short:   "Reboot the system",
	Example: "  speedrun system reboot",
	Args:    cobra.NoArgs,
	RunE:    reboot,
}

func init() {
	systemCmd.SetUsageTemplate(usage)
	systemCmd.AddCommand(rebootCmd)
}

func reboot(cmd *cobra.Command, _ []string) error {
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

			addr := net.JoinHostPort(portal.GetAddress(usePrivateIP), "1337")
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

			r, err := c.SystemReboot(ctx, &portalpb.SystemRebootRequest{})
			if err != nil {
				log.Error(err.Error())
				return
			}
			log.WithField("state", r.GetState()).Infof(r.GetMessage())

		})
	}
	pool.StopAndWait()
	return nil
}
