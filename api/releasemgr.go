package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alexfacciorusso/ghurlparse"
	"github.com/danesparza/appupgrade/dpkg"
	"github.com/danesparza/appupgrade/github"
	"github.com/gorilla/mux"
	"github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// VersionReport defines version information for a given package
type VersionReport struct {
	Name              string `json:"name"`             // The package name
	InstalledVersion  string `json:"installedversion"` // The current (installed) version of the package
	LatestVersion     string `json:"latestversion"`    // The latest available version of the package
	UpdateDownloadUrl string `json:"downloadurl"`      // The url to get the latest update
	UpgradeAvailable  bool   `json:"upgradeavailable"` // 'true' if there is an upgrade available
}

// GetVersionInfoForPackage gets the version information for the given package
func (service Service) GetVersionInfoForPackage(rw http.ResponseWriter, req *http.Request) {

	retval := VersionReport{}

	//	Parse the request
	vars := mux.Vars(req)

	//	Get the package name:
	packageName := vars["package"]
	retval.Name = packageName

	//	Log our request
	log.WithFields(log.Fields{
		"route":   req.URL.RequestURI(),
		"package": packageName,
	}).Debug("version info request")

	//	Get the list of packages (and their github urls)
	monitorPackages := viper.GetStringMap("packages")

	if packageRepo, packageIsMonitored := monitorPackages[packageName]; packageIsMonitored {
		//	Get currently installed package version
		currentVersion, err := dpkg.GetCurrentVersionForPackage(packageName)
		if err != nil {
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"package": packageName,
				}).Error("problem getting current version for package")
				sendErrorResponse(rw, fmt.Errorf("problem getting current version for package: %s", packageName), http.StatusInternalServerError)
				return
			}
		}

		retval.InstalledVersion = currentVersion
		log.WithFields(log.Fields{
			"package":        packageName,
			"currentVersion": currentVersion,
		}).Debug("Found current version")

		//	... parse the repo information
		valid, user, repo := ghurlparse.DestructureRepoURL(fmt.Sprintf("%s", packageRepo))
		if valid {
			releases, err := github.GetVersionsForRepo(user, repo)

			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"user": user,
					"repo": repo,
				}).Error("problem getting versions for repo")
				sendErrorResponse(rw, fmt.Errorf("problem getting versions for repo: %s/%s", user, repo), http.StatusFailedDependency)
				return
			}

			//	If we seem to have a list of releases, print the latest release information
			if len(releases) > 0 {
				retval.LatestVersion = releases[0].Version
				retval.UpdateDownloadUrl = releases[0].DownloadUrl

				log.WithFields(log.Fields{
					"package":    packageName,
					"version":    releases[0].Version,
					"releaseUrl": releases[0].DownloadUrl,
				}).Debug("Found latest release for package.")
			}

			//	See if the latest version is greater than the installed version.  If so, an update is available
			verInstalled, err := version.NewVersion(retval.InstalledVersion)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"user":           user,
					"repo":           repo,
					"package":        packageName,
					"currentversion": retval.InstalledVersion,
					"latestversion":  retval.LatestVersion,
				}).Error("failed to parse current version")
				sendErrorResponse(rw, fmt.Errorf("failed to parse current version: %s", retval.InstalledVersion), http.StatusInternalServerError)
				return
			}

			verLatest, err := version.NewVersion(retval.LatestVersion)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"user":           user,
					"repo":           repo,
					"package":        packageName,
					"currentversion": retval.InstalledVersion,
					"latestversion":  retval.LatestVersion,
				}).Error("failed to parse latest version")
				sendErrorResponse(rw, fmt.Errorf("failed to parse latest version: %s", retval.LatestVersion), http.StatusInternalServerError)
				return
			}

			if verLatest.GreaterThan(verInstalled) {
				retval.UpgradeAvailable = true
			}

		}
	} else {
		//	We're not monitoring the requested package:  Return an error
		sendErrorResponse(rw, fmt.Errorf("not monitoring the package %s", packageName), http.StatusNotFound)
		return
	}

	//	Log our found information
	log.WithFields(log.Fields{
		"route":       req.URL.RequestURI(),
		"versionInfo": retval,
	}).Debug("returning version response")

	//	Our return value
	response := SystemResponse{
		Message: "Version data fetched",
		Data:    retval,
	}

	//	Serialize to JSON & return the response:
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(rw).Encode(response)
}
