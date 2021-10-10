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
			const addr = "0.0.0.0:1337"

			m := drpcmux.New()
			portalpb.DRPCRegisterPortal(m, &portal.Server{})
			log.Infof("Started portal on %s", addr)

			s := drpcserver.New(m)

			tlsConfig, err := itls.GenerateTLSConfig()
			if err != nil {
				log.Fatalf("failed to generate tls config: %v", err)
			}

			lis, err := tls.Listen("tcp", addr, tlsConfig)
			if err != nil {
				log.Fatalf("failed to listen: %v", err)
			}
			defer lis.Close()

			ctx := context.Background()
			if err := s.Serve(ctx, lis); err != nil {
				log.Fatalf("failed to serve: %v", err)
			}
		},
	}

	dir := "/etc/portal"
	path := filepath.Join(dir, "config.toml")

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", path, "config file")
	rootCmd.PersistentFlags().StringP("loglevel", "l", "info", "Log level")
	viper.BindPFlag("loglevel", rootCmd.PersistentFlags().Lookup("loglevel"))

	rootCmd.DisableSuggestions = false

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}

func initConfig() {
	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Errorf("Couldn't read config: %s", viper.ConfigFileUsed())
	}

	lvl, err := log.ParseLevel(viper.GetString("loglevel"))
	if err != nil {
		log.Fatalf("Couldn't parse log level: %s (%s)", err, lvl)
		return
	}
	log.SetLevel(lvl)
}
