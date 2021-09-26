package cli

import (
	"context"
	"strings"

	"github.com/alitto/pond"
	"github.com/apex/log"
	"github.com/speedrunsh/speedrun/pkg/common/key"
	transport "github.com/speedrunsh/speedrun/pkg/common/transport"
	"github.com/speedrunsh/speedrun/pkg/speedrun/cloud"
	"github.com/speedrunsh/speedrun/pkg/speedrun/result"
	portalpb "github.com/speedrunsh/speedrun/proto/portal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	serviceCmd.PersistentFlags().Bool("only-failures", false, "Print only failures and errors")
	viper.BindPFlag("transport.insecure", serviceCmd.PersistentFlags().Lookup("insecure"))
	viper.BindPFlag("portal.use-tunnel", serviceCmd.PersistentFlags().Lookup("use-tunnel"))
	viper.BindPFlag("portal.use-private-ip", serviceCmd.PersistentFlags().Lookup("use-private-ip"))
	viper.BindPFlag("gcp.use-oslogin", serviceCmd.PersistentFlags().Lookup("use-oslogin"))
	viper.BindPFlag("portal.only-failures", serviceCmd.PersistentFlags().Lookup("only-failures"))
}

func action(cmd *cobra.Command, args []string) error {
	project := viper.GetString("gcp.projectid")
	useTunnel := viper.GetBool("portal.use-tunnel")
	insecure := viper.GetBool("transport.insecure")
	usePrivateIP := viper.GetBool("portal.use-private-ip")
	useOSlogin := viper.GetBool("gcp.use-oslogin")
	onlyFailures := viper.GetBool("portal.only-failures")

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

	pool := pond.New(1000, 10000)
	res := result.NewResult()

	for _, i := range instances {
		instance := i
		pool.Submit(func() {
			t, err := transport.NewGRPCTransport(instance.Address, transport.WithSSH(useTunnel), transport.WithSSHKey(*k), transport.WithInsecure(insecure))
			if err != nil {
				res.AddError(instance.Name, err)
				return
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
					res.AddFailure(instance.Name, e.Message())
				}
				return
			}
			res.AddSuccess(instance.Name, r.GetContent())
		})
	}
	pool.StopAndWait()
	res.Print(onlyFailures)

	return nil
}
