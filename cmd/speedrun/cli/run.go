package cli

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/alitto/pond"
	itls "github.com/speedrunsh/speedrun/pkg/common/tls"
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
	runCmd.Flags().Bool("insecure", true, "Skip Portal's certificate verification (gRPC/QUIC)")
	runCmd.Flags().Bool("use-private-ip", false, "Connect to private IPs instead of public ones")
	viper.BindPFlag("gcp.projectid", runCmd.Flags().Lookup("projectid"))
	viper.BindPFlag("transport.insecure", runCmd.Flags().Lookup("insecure"))
	viper.BindPFlag("portal.use-private-ip", runCmd.Flags().Lookup("use-private-ip"))

}

func run(cmd *cobra.Command, args []string) error {
	command := strings.Join(args, " ")
	s := strings.Split(command, " ")
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

			r, err := c.RunCommand(ctx, &portalpb.Command{Name: s[0], Args: s[1:]})
			if err != nil {
				log.WithFields(fields).Warn(err.Error())
				return
			}
			log.WithFields(fields).Infof(r.GetContent())
		})
	}
	pool.StopAndWait()
	return nil
}
