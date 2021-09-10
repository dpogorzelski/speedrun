package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/apex/log"
	"github.com/speedrunsh/portal"
	"github.com/speedrunsh/speedrun/cloud"
	"github.com/speedrunsh/speedrun/colors"
	"github.com/speedrunsh/speedrun/key"
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
	Short:   "restart a service",
	Example: "  speedrun service restart nginx",
	Args:    cobra.MinimumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
	RunE: action,
}

var startCmd = &cobra.Command{
	Use:     "start <servicename>",
	Short:   "start a service",
	Example: "  speedrun service start nginx",
	Args:    cobra.MinimumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
	RunE: action,
}

var stopCmd = &cobra.Command{
	Use:     "stop <servicename>",
	Short:   "stop a service",
	Example: "  speedrun service stop nginx",
	Args:    cobra.MinimumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
	},
	RunE: action,
}

func init() {
	restartCmd.Flags().StringP("target", "t", "", "Select instances that match the given criteria")
	restartCmd.Flags().Bool("ignore-fingerprint", false, "Ignore host's fingerprint mismatch")
	serviceCmd.PersistentFlags().Bool("use-tunnel", true, "Connect to the portals via SSH tunnel")
	viper.BindPFlag("ssh.ignore-fingerprint", serviceCmd.Flags().Lookup("ignore-fingerprint"))
	viper.BindPFlag("portal.use-tunnel", serviceCmd.Flags().Lookup("use-tunnel"))
	serviceCmd.SetUsageTemplate(usage)
	serviceCmd.AddCommand(restartCmd)
	serviceCmd.AddCommand(startCmd)
	serviceCmd.AddCommand(stopCmd)
}

func action(cmd *cobra.Command, args []string) error {
	project := viper.GetString("gcp.projectid")
	useTunnel := viper.GetBool("portal.use-tunnel")
	ignoreFingerprint := viper.GetBool("ssh.ignore-fingerprint")

	target, err := cmd.Flags().GetString("target")
	if err != nil {
		return err
	}

	gcpClient, err := cloud.NewGCPClient(project)
	if err != nil {
		return err
	}

	log.Info("Fetching instance list")
	instances, err := gcpClient.GetInstances(target, false)
	if err != nil {
		return err
	}

	if len(instances) == 0 {
		log.Warn("No instances found")
		return nil
	}

	var k *key.Key
	if useTunnel {
		path, err := key.Path()
		if err != nil {
			return err
		}

		k, err = key.Read(path)
		if err != nil {
			return err
		}
	}

	for _, instance := range instances {
		var t *grpc.ClientConn
		var err error
		if useTunnel {
			t, err = portal.NewTransport(instance.Address, portal.WithSSH(*k), portal.WithInsecure(ignoreFingerprint))
		} else {
			t, err = portal.NewTransport(instance.Address)
		}
		if err != nil {
			log.Warn(err.Error())
			continue
		}
		defer t.Close()
		c := portal.NewPortalClient(t)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var r *portal.Response
		switch cmd.Name() {
		case "restart":
			r, err = c.ServiceRestart(ctx, &portal.Service{Name: strings.Join(args, " ")})
		case "start":
			r, err = c.ServiceStart(ctx, &portal.Service{Name: strings.Join(args, " ")})
		case "stop":
			r, err = c.ServiceStop(ctx, &portal.Service{Name: strings.Join(args, " ")})
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
