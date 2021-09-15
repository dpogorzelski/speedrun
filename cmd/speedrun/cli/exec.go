package cli

// var runCmd = &cobra.Command{
// 	Use:     "run <command to run>",
// 	Short:   "Run a command on remote servers",
// 	Example: "  speedrun run whoami\n  speedrun run whoami --only-failures --target \"labels.foo = bar AND labels.environment = staging\"",
// 	Args:    cobra.MinimumNArgs(1),
// 	PreRun: func(cmd *cobra.Command, args []string) {
// 		initConfig()
// 	},
// 	RunE: run,
// }

// func init() {
// 	runCmd.Flags().StringP("target", "t", "", "Fetch instances that match the target selection criteria")
// 	runCmd.Flags().String("projectid", "", "Override GCP project id")
// 	runCmd.Flags().Bool("only-failures", false, "Print only failures and errors")
// 	runCmd.Flags().Bool("ignore-fingerprint", false, "Ignore host's fingerprint mismatch")
// 	runCmd.Flags().Duration("timeout", time.Duration(10*time.Second), "SSH connection timeout")
// 	runCmd.Flags().Int("concurrency", 100, "Number of maximum concurrent SSH workers")
// 	runCmd.Flags().Bool("use-private-ip", false, "Connect to private IPs instead of public ones")
// 	runCmd.Flags().Bool("use-oslogin", false, "Authenticate via OS Login")
// 	runCmd.Flags().Bool("use-tunnel", true, "Connect to the portals via SSH tunnel")
// 	viper.BindPFlag("gcp.projectid", runCmd.Flags().Lookup("projectid"))
// 	viper.BindPFlag("gcp.use-oslogin", runCmd.Flags().Lookup("use-oslogin"))
// 	viper.BindPFlag("ssh.timeout", runCmd.Flags().Lookup("timeout"))
// 	viper.BindPFlag("ssh.ignore-fingerprint", runCmd.Flags().Lookup("ignore-fingerprint"))
// 	viper.BindPFlag("ssh.only-failures", runCmd.Flags().Lookup("only-failures"))
// 	viper.BindPFlag("ssh.concurrency", runCmd.Flags().Lookup("concurrency"))
// 	viper.BindPFlag("ssh.use-private-ip", runCmd.Flags().Lookup("use-private-ip"))
// 	viper.BindPFlag("portal.use-tunnel", runCmd.Flags().Lookup("use-tunnel"))
// 	runCmd.SetUsageTemplate(usage)
// }

// func run(cmd *cobra.Command, args []string) error {
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

// 	var k *key.Key
// 	if useTunnel {
// 		path, err := key.Path()
// 		if err != nil {
// 			return err
// 		}

// 		k, err = key.Read(path)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	for _, instance := range instances {
// 		var t *grpc.ClientConn
// 		var err error
// 		if useTunnel {
// 			t, err = portal.NewTransport(instance.Address, portal.WithSSH(*k), portal.WithInsecure(ignoreFingerprint))
// 		} else {
// 			t, err = portal.NewTransport(instance.Address)
// 		}
// 		if err != nil {
// 			log.Warn(err.Error())
// 			continue
// 		}
// 		defer t.Close()
// 		c := portal.NewPortalClient(t)

// 		ctx, cancel := context.WithCancel(context.Background())
// 		defer cancel()

// 		r, err := c.RunCommand(ctx, &portal.Command{Name: strings.Join(args, " ")})
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
