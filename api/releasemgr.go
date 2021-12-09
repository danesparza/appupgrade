package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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

// GetVersionInfoForPackage godoc
// @Summary gets the version information for the given package
// @Description gets the version information for the given package
// @Tags package
// @Accept  json
// @Produce  json
// @Param package path string true "The package to get information for"
// @Success 200 {object} api.SystemResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 424 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /package/{package}/info [get]
func (service Service) GetVersionInfoForPackage(rw http.ResponseWriter, req *http.Request) {

	retval := VersionReport{}

	//	Parse the request
	vars := mux.Vars(req)

	//	Get the package name:
	packageName := vars["package"]
	if strings.TrimSpace(packageName) == "" {
		sendErrorResponse(rw, fmt.Errorf("package is a required parameter and should not be blank"), http.StatusBadRequest)
	}

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

// UpdatePackageToVersion godoc
// @Summary updates a package to the specified version
// @Description updates a package to the specified version
// @Tags package
// @Accept  json
// @Produce  json
// @Param package path string true "The package to update"
// @Param version path string true "The version to update to"
// @Success 200 {object} api.SystemResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 424 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /package/{package}/updatetoversion/{version} [post]
func (service Service) UpdatePackageToVersion(rw http.ResponseWriter, req *http.Request) {
	retval := ""

	//	Parse the request
	vars := mux.Vars(req)

	//	Get the package name:
	packageName := vars["package"]
	if strings.TrimSpace(packageName) == "" {
		sendErrorResponse(rw, fmt.Errorf("package is a required parameter and should not be blank"), http.StatusBadRequest)
	}

	reqVersion := vars["version"]
	if strings.TrimSpace(reqVersion) == "" {
		sendErrorResponse(rw, fmt.Errorf("version is a required parameter and should not be blank"), http.StatusBadRequest)
	}

	versionRequested, err := version.NewVersion(reqVersion)
	if err != nil {
		sendErrorResponse(rw, fmt.Errorf("version is a required parameter and should be in a format similar to v1.23"), http.StatusBadRequest)
	}

	//	Log our request
	log.WithFields(log.Fields{
		"route":   req.URL.RequestURI(),
		"package": packageName,
		"version": reqVersion,
	}).Debug("package update request")

	//	Get the list of packages (and their github urls)
	monitorPackages := viper.GetStringMap("packages")

	//	Make sure the requested package is being monitored ...
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
					"user":    user,
					"repo":    repo,
					"package": packageName,
				}).Error("problem getting versions for repo")
				sendErrorResponse(rw, fmt.Errorf("problem getting versions for repo: %s/%s", user, repo), http.StatusFailedDependency)
				return
			}

			//	Look for the requested version in the list of versions
			for _, release := range releases {
				versionFound, err := version.NewVersion(release.Version)
				if err != nil {
					log.WithError(err).WithFields(log.Fields{
						"user":           user,
						"repo":           repo,
						"package":        packageName,
						"releaseVersion": release.Version,
					}).Warn("problem parsing a found release version - skipping to next version")
					continue
				}

				//	If we found the version...
				if versionRequested.Equal(versionFound) {
					log.WithFields(log.Fields{
						"user":        user,
						"repo":        repo,
						"package":     packageName,
						"version":     release.Version,
						"downloadurl": release.DownloadUrl,
					}).Debug("found requested version information")

					//	Download the file
					packageFile, err := github.DownloadFile(release.DownloadUrl)
					if err != nil {
						log.WithError(err).WithFields(log.Fields{
							"user":           user,
							"repo":           repo,
							"package":        packageName,
							"releaseVersion": release.Version,
							"downloadurl":    release.DownloadUrl,
						}).Error("problem downloading the package file for release")
						sendErrorResponse(rw, fmt.Errorf("problem downloading the package file for release: %s", release.DownloadUrl), http.StatusInternalServerError)
						return
					}

					//	Remove the previous package
					rpkgOut, err := dpkg.RemovePackage(packageName)
					if err != nil {
						log.WithError(err).WithFields(log.Fields{
							"package": packageName,
						}).Error("problem removing the old package")
						sendErrorResponse(rw, fmt.Errorf("problem removing the package: %s", packageName), http.StatusInternalServerError)
						return
					}

					log.WithFields(log.Fields{
						"package": packageName,
						"output":  rpkgOut,
					}).Debug("Removed package")

					//	Install the new package
					ipkgOut, err := dpkg.InstallPackage(packageFile)
					if err != nil {
						log.WithError(err).WithFields(log.Fields{
							"package":     packageName,
							"packageFile": packageFile,
						}).Error("problem installing the package")
						sendErrorResponse(rw, fmt.Errorf("problem installing the package: %s", packageFile), http.StatusInternalServerError)
						return
					}

					log.WithFields(log.Fields{
						"package": packageName,
						"output":  ipkgOut,
					}).Debug("Installed package")

					retval = fmt.Sprintf("Installed %s version %s", packageName, reqVersion)
				}
			}
		}
	} else {
		//	We're not monitoring the requested package:  Return an error
		sendErrorResponse(rw, fmt.Errorf("not monitoring the package %s", packageName), http.StatusNotFound)
		return
	}

	//	Log our found information
	log.WithFields(log.Fields{
		"route":  req.URL.RequestURI(),
		"retval": retval,
	}).Debug("returning update package response")

	//	Our return value
	response := SystemResponse{
		Message: "Package updated",
		Data:    retval,
	}

	//	Serialize to JSON & return the response:
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(rw).Encode(response)
}
