package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: start,
}

func start(cmd *cobra.Command, args []string) {
	//	If we have a config file, report it:
	if viper.ConfigFileUsed() != "" {
		log.Debugf("Using config file: %s", viper.ConfigFileUsed())
	} else {
		log.Debug("No config file found.")
	}

	//	Emit what we know:
	monitorPackages := viper.GetStringMap("packages")
	for k, v := range monitorPackages {
		fmt.Printf("%s -> %s\n", k, v)
	}

	//	We could use https://github.com/alexfacciorusso/ghurlparse to parse the github urls
}

func init() {
	rootCmd.AddCommand(startCmd)
}
