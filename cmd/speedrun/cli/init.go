package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize speedrun",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := filepath.Dir(cfgFile)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.Mkdir(dir, 0755)
			if err != nil {
				return err
			}
		}

		err := viper.SafeWriteConfigAs(cfgFile)
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

func init() {
	initCmd.SetUsageTemplate(usage)
}
