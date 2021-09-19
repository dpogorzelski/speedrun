package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/apex/log"
	"github.com/speedrunsh/speedrun/pkg/common/colors"
	"github.com/speedrunsh/speedrun/pkg/common/key"
	transport "github.com/speedrunsh/speedrun/pkg/common/transport"
	"github.com/speedrunsh/speedrun/pkg/speedrun/cloud"
	portalpb "github.com/speedrunsh/speedrun/proto/portal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
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
	serviceCmd.PersistentFlags().Bool("insecure", true, "Ignore host's fingerprint mismatch (SSH) or skip Portal's certificate verification (gRPC/QUIC)")
	serviceCmd.PersistentFlags().Bool("use-tunnel", true, "Connect to the portals via SSH tunnel")
	serviceCmd.PersistentFlags().Bool("use-private-ip", false, "Connect to private IPs instead of public ones")
	serviceCmd.PersistentFlags().Bool("use-oslogin", false, "Authenticate via OS Login")
	viper.BindPFlag("transport.insecure", serviceCmd.PersistentFlags().Lookup("insecure"))
	viper.BindPFlag("portal.use-tunnel", serviceCmd.PersistentFlags().Lookup("use-tunnel"))
	viper.BindPFlag("portal.use-private-ip", serviceCmd.PersistentFlags().Lookup("use-private-ip"))
	viper.BindPFlag("gcp.use-oslogin", serviceCmd.PersistentFlags().Lookup("use-oslogin"))
}

func action(cmd *cobra.Command, args []string) error {
	project := viper.GetString("gcp.projectid")
	useTunnel := viper.GetBool("portal.use-tunnel")
	insecure := viper.GetBool("transport.insecure")
	usePrivateIP := viper.GetBool("portal.use-private-ip")
	useOSlogin := viper.GetBool("gcp.use-oslogin")

	target, err := cmd.Flags().GetString("target")
	if err != nil {
		return err
	}

	gcpClient, err := cloud.NewGCPClient(project)
	if err != nil {
		return err
	}

	log.Info("Fetching instance list")
	instances, err := gcpClient.GetInstances(target, usePrivateIP)
	if err != nil {
		return err
	}

	if len(instances) == 0 {
		log.Warn("No instances found")
		return nil
	}

	var k *key.Key
	if useTunnel {
		log.Debug("Using SSH tunnel")
		path, err := key.Path()
		if err != nil {
			return err
		}

		k, err = key.Read(path)
		if err != nil {
			return err
		}

		if useOSlogin {
			log.Debug("Using OS Login")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			user, err := gcpClient.GetSAUsername(ctx)
			if err != nil {
				return err
			}
			k.User = user
		}
	}

	for _, instance := range instances {
		var t *grpc.ClientConn
		var err error
		if useTunnel {
			t, err = transport.NewTransport(instance.Address, transport.WithSSH(*k), transport.WithInsecure(insecure))
		} else {
			t, err = transport.NewTransport(instance.Address)
		}
		if err != nil {
			log.Warn(err.Error())
			continue
		}
		defer t.Close()
		c := portalpb.NewPortalClient(t)

		ctx, cancel := context.WithCancel(context.Background())
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
			if e, ok := status.FromError(err); ok {
				fmt.Printf("  %s:\n    %s\n\n", colors.Yellow(instance.Name), e.Message())
			}
			continue
		}
		fmt.Printf("  %s:\n    %s\n\n", colors.Green(instance.Name), r.GetContent())
	}

	return nil
}
