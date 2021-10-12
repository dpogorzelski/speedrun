package cli

import (
	"context"
	"crypto/tls"
	"fmt"
	"path/filepath"

	"github.com/apex/log"
	itls "github.com/speedrunsh/speedrun/pkg/common/tls"
	"github.com/speedrunsh/speedrun/pkg/portal"
	portalpb "github.com/speedrunsh/speedrun/proto/portal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"storj.io/drpc/drpcmux"
	"storj.io/drpc/drpcserver"
)

var cfgFile string
var version string
var commit string
var date string

func Execute() {
	var rootCmd = &cobra.Command{
		Use:           "portal",
		Short:         "Control your compute fleet at scale",
		Version:       fmt.Sprintf("%s, commit: %s, date: %s", version, commit, date),
		SilenceUsage:  false,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			m := drpcmux.New()
			portalpb.DRPCRegisterPortal(m, &portal.Server{})
			s := drpcserver.New(m)

			tlsConfig, err := itls.GenerateTLSConfig()
			if err != nil {
				log.Fatalf("failed to generate tls config: %v", err)
			}

			port := viper.GetInt("port")
			ip := viper.GetString("address")
			addr := fmt.Sprintf("%s:%d", ip, port)
			lis, err := tls.Listen("tcp", addr, tlsConfig)
			if err != nil {
				log.Fatalf("failed to listen: %v", err)
			}
			defer lis.Close()

			ctx := context.Background()
			log.Infof("Started portal on %s", addr)
			if err := s.Serve(ctx, lis); err != nil {
				log.Fatalf("failed to serve: %v", err)
			}
		},
	}

	dir := "/etc/portal"
	path := filepath.Join(dir, "config.toml")

	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(initCmd)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", path, "config file")
	rootCmd.PersistentFlags().StringP("loglevel", "l", "info", "Log level")
	rootCmd.Flags().IntP("port", "p", 1337, "Port to listen on for connections")
	rootCmd.Flags().StringP("address", "a", "0.0.0.0", "Address to listen on for connections")
	viper.BindPFlag("loglevel", rootCmd.PersistentFlags().Lookup("loglevel"))
	viper.BindPFlag("port", rootCmd.Flags().Lookup("port"))
	viper.BindPFlag("address", rootCmd.Flags().Lookup("address"))

	rootCmd.DisableSuggestions = false

	if err := rootCmd.Execute(); err != nil {
		log.Error(err.Error())
	}
}

func initConfig() {
	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Warnf("Couldn't read config at \"%s\", starting with default settings", viper.ConfigFileUsed())
	}

	lvl, err := log.ParseLevel(viper.GetString("loglevel"))
	if err != nil {
		log.Fatalf("couldn't parse log level: %s (%s)", err, lvl)
		return
	}
	log.SetLevel(lvl)
}
