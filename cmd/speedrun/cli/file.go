package cli

import (
	"context"
	"crypto/tls"
	"net"
	"os"
	"strings"
	"time"

	"github.com/alitto/pond"
	"github.com/apex/log"
	"github.com/dpogorzelski/speedrun/pkg/speedrun/cloud"
	portalpb "github.com/dpogorzelski/speedrun/proto/portal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"storj.io/drpc/drpcconn"
)

var fileCmd = &cobra.Command{
	Use:              "file",
	Short:            "Manage files",
	TraverseChildren: true,
}

var readCmd = &cobra.Command{
	Use:     "read <path>",
	Short:   "Read a file",
	Example: "  speedrun file read /etc/resolv.conf",
	Args:    cobra.MinimumNArgs(1),
	RunE:    read,
}

var cpCmd = &cobra.Command{
	Use:     "cp <src> <dst>",
	Short:   "Copy a file",
	Example: "  speedrun file cp myfile /tmp/myfile",
	Args:    cobra.MinimumNArgs(1),
	RunE:    cp,
}

func init() {
	fileCmd.SetUsageTemplate(usage)
	fileCmd.AddCommand(readCmd)
	fileCmd.AddCommand(cpCmd)
}

func read(cmd *cobra.Command, args []string) error {
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

			path := strings.Join(args, " ")
			r, err := c.FileRead(ctx, &portalpb.FileReadRequest{Path: path})
			if err != nil {
				log.Error(err.Error())
				return
			}
			log.WithField("state", r.GetState()).Infof("Contents of %s:", path)
			println(r.GetContent())

		})
	}
	pool.StopAndWait()
	return nil
}

func cp(cmd *cobra.Command, args []string) error {
	usePrivateIP := viper.GetBool("portal.use-private-ip")

	tlsConfig, err := cloud.SetupTLS()
	if err != nil {
		return err
	}

	target, err := cmd.Flags().GetString("target")
	if err != nil {
		return err
	}

	remoteSrc := strings.HasPrefix(args[0], ":")
	remoteDst := strings.HasPrefix(args[1], ":")

	var content []byte
	if !remoteSrc {
		content, err = os.ReadFile(args[0])
		if err != nil {
			return err
		}
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

			var r *portalpb.FileCpResponse

			r, err = c.FileCp(ctx, &portalpb.FileCpRequest{Src: strings.TrimPrefix(args[0], ":"), Dst: strings.TrimPrefix(args[1], ":"), Content: content, RemoteSrc: remoteSrc, RemoteDst: remoteDst})
			if err != nil {
				log.Error(err.Error())
				return
			}
			if len(r.GetContent()) > 0 {
				err := os.WriteFile(args[1], r.GetContent(), 0644)
				if err != nil {
					log.Error(err.Error())
					return
				}
			}

			log.WithField("state", r.GetState()).Infof("Done")
		})
	}
	pool.StopAndWait()
	return nil
}
