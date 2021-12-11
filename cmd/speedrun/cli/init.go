package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	initCmd.SetUsageTemplate(usage)
	initCmd.Flags().BoolP("print", "p", false, "Print default config to stdout")
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize speedrun",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.SetDefault("gcp.projectid", "")
		print, err := cmd.Flags().GetBool("print")
		if err != nil {
			return err
		}

		if print {
			c := viper.AllSettings()
			bs, err := toml.Marshal(c)
			if err != nil {
				return fmt.Errorf("unable to marshal config: %v", err)
			}

			fmt.Println(string(bs))
			return nil
		}

		dir := filepath.Dir(cfgFile)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.Mkdir(dir, 0755)
			if err != nil {
				return err
			}
		}

		err = viper.SafeWriteConfigAs(cfgFile)
		if err != nil {
			if _, ok := err.(viper.ConfigFileAlreadyExistsError); !ok {
				return fmt.Errorf("couldn't save config at \"%s\" (%s)", viper.ConfigFileUsed(), err)
			}
		} else {
			log.Infof("Your config was saved at \"%s\"", viper.ConfigFileUsed())
			return nil
		}

		log.Infof("Config already exists at \"%s\", no changes applied", viper.ConfigFileUsed())
		return nil
	},
}
