package cmd

import (
	"fmt"

	"github.com/alexfacciorusso/ghurlparse"
	"github.com/danesparza/appupgrade/dpkg"
	"github.com/danesparza/appupgrade/github"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the server",
	Long:  `The start command starts the app upgrade server`,
	Run:   start,
}

func start(cmd *cobra.Command, args []string) {
	//	If we have a config file, report it:
	if viper.ConfigFileUsed() != "" {
		log.Debugf("Using config file: %s", viper.ConfigFileUsed())
	} else {
		log.Debug("No config file found.")
	}

	//	Get the list of packages (and their github urls)
	monitorPackages := viper.GetStringMap("packages")

	//	Emit what we know:
	log.WithFields(log.Fields{
		"Monitor packages": monitorPackages,
	}).Info("Starting up")

	//	For each package ...
	for k, v := range monitorPackages {
		//	Get currently installed package version
		currentVersion, err := dpkg.GetCurrentVersionForPackage(k)
		if err != nil {
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"package": k,
				}).Error("problem getting current version for package")
				return
			}
		}

		log.WithFields(log.Fields{
			"package":        k,
			"currentVersion": currentVersion,
		}).Info("Found current version")

		//	... parse the repo information
		valid, user, repo := ghurlparse.DestructureRepoURL(fmt.Sprintf("%s", v))
		if valid {
			releases, err := github.GetVersionsForRepo(user, repo)

			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"user": user,
					"repo": repo,
				}).Error("problem getting versions for repo")
				return
			}

			//	If we seem to have a list of releases, print the latest release information
			if len(releases) > 0 {
				log.WithFields(log.Fields{
					"package":    k,
					"version":    releases[0].Version,
					"releaseUrl": releases[0].DownloadUrl,
				}).Info("Found latest release for package.")
			}

		}
	}

}

func init() {
	rootCmd.AddCommand(startCmd)
}
