package cli

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	jsonhandler "github.com/apex/log/handlers/json"
	texthandler "github.com/apex/log/handlers/text"
	"github.com/speedrunsh/speedrun/pkg/common/cryptoutil"
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
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			insecure := viper.GetBool("tls.insecure")
			caPath := viper.GetString("tls.ca")
			certPath := viper.GetString("tls.cert")
			keyPath := viper.GetString("tls.key")

			m := drpcmux.New()
			var err error
			err = portalpb.DRPCRegisterPortal(m, &portal.Server{})
			if err != nil {
				return fmt.Errorf("could not register DRPC server: %v", err)
			}
			s := drpcserver.New(m)

			var tlsConfig *tls.Config
			if insecure {
				log.Warn("Using insecure TLS configuration, this should be avoided in production environments")
				tlsConfig, err = cryptoutil.InsecureTLSConfig()
			} else {
				tlsConfig, err = cryptoutil.ServerTLSConfig(caPath, certPath, keyPath)
			}
			if err != nil {
				return fmt.Errorf("could not instantiate TLS config: %v", err)
			}

			port := viper.GetInt("port")
			ip := viper.GetString("address")
			addr := fmt.Sprintf("%s:%d", ip, port)
			lis, err := tls.Listen("tcp", addr, tlsConfig)
			if err != nil {
				return fmt.Errorf("could not create a TCP socket on \"%s\": %v", addr, err)
			}
			defer lis.Close()

			ctx := context.Background()
			log.Infof("Starting portal on %s", addr)
			return s.Serve(ctx, lis)
		},
	}

	dir := "/etc/portal"
	configPath := filepath.Join(dir, "config.toml")

	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(initCmd)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", configPath, "config file")
	rootCmd.PersistentFlags().StringP("loglevel", "l", "info", "Log level")
	rootCmd.PersistentFlags().BoolP("json", "j", false, "Output logs in JSON format")
	rootCmd.Flags().IntP("port", "p", 1337, "Port to listen on for connections")
	rootCmd.Flags().StringP("address", "a", "0.0.0.0", "Address to listen on for connections")
	rootCmd.Flags().Bool("insecure", false, "Skip client certificate verification")
	rootCmd.Flags().String("ca", "ca.crt", "Path to the CA cert")
	rootCmd.Flags().String("cert", "portal.crt", "Path to the server cert")
	rootCmd.Flags().String("key", "portal.key", "Path to the server key")

	viper.BindPFlag("tls.insecure", rootCmd.Flags().Lookup("insecure"))
	viper.BindPFlag("tls.ca", rootCmd.Flags().Lookup("ca"))
	viper.BindPFlag("tls.cert", rootCmd.Flags().Lookup("cert"))
	viper.BindPFlag("tls.key", rootCmd.Flags().Lookup("key"))
	viper.BindPFlag("logging.loglevel", rootCmd.PersistentFlags().Lookup("loglevel"))
	viper.BindPFlag("logging.json", rootCmd.PersistentFlags().Lookup("json"))
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

	json := viper.GetBool("logging.json")
	if json {
		handler := jsonhandler.New(os.Stdout)
		log.SetHandler(handler)
	} else {
		handler := texthandler.New(os.Stdout)
		log.SetHandler(handler)
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Warnf("Couldn't read config at \"%s\", starting with default settings", viper.ConfigFileUsed())
	}

	lvl, err := log.ParseLevel(viper.GetString("logging.loglevel"))
	if err != nil {
		log.Fatalf("couldn't parse log level: %s (%s)", err, lvl)
		return
	}
	log.SetLevel(lvl)
}
