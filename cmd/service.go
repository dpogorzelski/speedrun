package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serviceCmd = &cobra.Command{
	Use:              "service",
	Short:            "Manage services",
	TraverseChildren: true,
}

var restartCmd = &cobra.Command{
	Use:     "restart <servicename>",
	Short:   "restart a service",
	Example: "  speedrun service restart nginx",
	Args:    cobra.MinimumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
	RunE: restart,
}

func init() {
	restartCmd.Flags().StringP("target", "t", "", "Select instances that match the given criteria")
	restartCmd.Flags().Bool("ignore-fingerprint", false, "Ignore host's fingerprint mismatch")
	serviceCmd.Flags().Bool("use-tunnel", true, "Connect to the portals via SSH tunnel")
	viper.BindPFlag("ssh.ignore-fingerprint", serviceCmd.Flags().Lookup("ignore-fingerprint"))
	viper.BindPFlag("portal.use-tunnel", serviceCmd.Flags().Lookup("use-tunnel"))

	serviceCmd.AddCommand(restartCmd)
}

// func restart(cmd *cobra.Command, args []string) error {
// 	serviceName := args[0]
// 	log.Debugf("Command: %s", serviceName)

// 	project := viper.GetString("gcp.projectid")
// 	useTunnel := viper.GetBool("portal.use-tunnel")
// 	ignoreFingerprint := viper.GetBool("ssh.ignore-fingerprint")

// 	target, err := cmd.Flags().GetString("target")
// 	if err != nil {
// 		return err
// 	}

// 	gcpClient, err := cloud.NewGCPClient(project)
// 	if err != nil {
// 		return err
// 	}

// 	log.Info("Fetching instance list")
// 	instances, err := gcpClient.GetInstances(target, false)
// 	if err != nil {
// 		return err
// 	}

// 	if len(instances) == 0 {
// 		log.Warn("No instances found")
// 		return nil
// 	}

// 	for _, instance := range instances {
// 		var grpcConn *grpc.ClientConn

// 		if useTunnel {
// 			path, err := key.Path()
// 			if err != nil {
// 				return err
// 			}

// 			k, err := key.Read(path)
// 			if err != nil {
// 				return err
// 			}

// 			log.WithField("instance", instance.Name).Debug("Using tunnel")
// 			if ignoreFingerprint {
// 				grpcConn, err = transport.SSHTransportInsecure(instance.Address, k)
// 			} else {
// 				grpcConn, err = transport.SSHTransport(instance.Address, k)
// 			}
// 			if err != nil {
// 				log.Errorf("%s:\n    %v\n", colors.Red(instance.Name), err)
// 				continue
// 			}
// 		} else {
// 			log.WithField("instance", instance.Name).Debug("Not using tunnel")
// 			grpcConn, err = transport.HTTP2Transport(instance.Address)
// 			if err != nil {
// 				log.Errorf("%s:\n    %v\n", colors.Red(instance.Name), err)
// 				continue
// 			}
// 		}

// 		defer grpcConn.Close()
// 		c := pcommand.NewPortalClient(grpcConn)

// 		ctx, cancel := context.WithCancel(context.Background())
// 		defer cancel()

// 		r, err := c.ServiceRestart(ctx, &command.req{Name: serviceName})
// 		if err != nil {

// 			if e, ok := status.FromError(err); ok {
// 				fmt.Printf("  %s:\n    %s\n\n", colors.Yellow(instance.Name), e.Message())
// 			}
// 			continue
// 		}
// 		fmt.Printf("  %s:\n    %s\n\n", colors.Green(instance.Name), r.GetContent())
// 	}

// 	return nil
// }

func restart(cmd *cobra.Command, args []string) error {

	return nil
}
