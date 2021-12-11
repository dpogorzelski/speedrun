package cli

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/alitto/pond"
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
	Example: "  speedrun run whoami\n  speedrun run whoami --target \"labels.foo = bar AND labels.environment = staging\"",
	Args:    cobra.MinimumNArgs(1),
	RunE:    run,
}

func init() {
	runCmd.SetUsageTemplate(usage)
}

func run(cmd *cobra.Command, args []string) error {
	command := strings.Join(args, " ")
	s := strings.Split(command, " ")
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
