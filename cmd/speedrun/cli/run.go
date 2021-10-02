package cli

import (
	"strings"

	"github.com/speedrunsh/speedrun/pkg/speedrun/cloud"
	"github.com/speedrunsh/speedrun/pkg/speedrun/result"

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
	runCmd.Flags().Bool("only-failures", false, "Print only failures and errors")
	runCmd.Flags().Bool("insecure", true, "Skip Portal's certificate verification (gRPC/QUIC)")
	runCmd.Flags().Bool("use-private-ip", false, "Connect to private IPs instead of public ones")
	viper.BindPFlag("gcp.projectid", runCmd.Flags().Lookup("projectid"))
	viper.BindPFlag("transport.insecure", runCmd.Flags().Lookup("insecure"))
	viper.BindPFlag("portal.only-failures", runCmd.Flags().Lookup("only-failures"))
	viper.BindPFlag("portal.use-private-ip", runCmd.Flags().Lookup("use-private-ip"))

}

func run(cmd *cobra.Command, args []string) error {
	command := strings.Join(args, " ")
	project := viper.GetString("gcp.projectid")
	insecure := viper.GetBool("transport.insecure")
	onlyFailures := viper.GetBool("portal.only-failures")
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

	var res *result.Result
	if insecure {
		log.Warn(command)
		log.Warnf("%s", usePrivateIP)
	} else {
		log.Warn(command)
	}

	res.Print(onlyFailures)
	return nil
}
