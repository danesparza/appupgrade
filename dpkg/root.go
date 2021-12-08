package dpkg

import (
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

// GetCurrentVersionForPackage returns the current installed version for a given package (or an error if it doesn't exist)
func GetCurrentVersionForPackage(packageName string) (string, error) {
	retval := ""

	log.WithFields(log.Fields{
		"package": packageName,
	}).Debug("requested current version for package")

	// Get the current installed version of the given unit
	// From https://askubuntu.com/a/712202/379868
	// dpkg-query --showformat='${Version}' --show python3-lxml
	versionInfo, err := exec.Command("dpkg-query", "--showformat", "${Version}", "--show", packageName).CombinedOutput()

	if err != nil {
		log.WithError(err).Error("problem running dpkg-query command")
		return retval, err
	}

	//	Remove leading/trailing whitespace if it exists:
	retval = strings.TrimSpace(string(versionInfo))

	log.WithFields(log.Fields{
		"package":        packageName,
		"currentVersion": retval,
	}).Debug("found current version for package")

	return retval, nil
}

// RemovePackage removes the given package
func RemovePackage(packageName string) (string, error) {
	retval := ""

	log.WithFields(log.Fields{
		"package": packageName,
	}).Debug("requested package remove")

	cmdOutput, err := exec.Command("dpkg", "-r", packageName).CombinedOutput()

	if err != nil {
		log.WithError(err).Error("problem running dpkg remove")
		return retval, err
	}

	//	Remove leading/trailing whitespace if it exists:
	retval = strings.TrimSpace(string(cmdOutput))

	log.WithFields(log.Fields{
		"package":   packageName,
		"cmdOutput": cmdOutput,
	}).Debug("removed package")

	return retval, nil
}

// InstallPackage installs the given deb file at the package path
func InstallPackage(packagePath string) (string, error) {
	retval := ""

	log.WithFields(log.Fields{
		"package": packagePath,
	}).Debug("requested package installation")

	cmdOutput, err := exec.Command("dpkg", "-i", packagePath).CombinedOutput()

	if err != nil {
		log.WithError(err).Error("problem running dpkg install")
		return retval, err
	}

	//	Remove leading/trailing whitespace if it exists:
	retval = strings.TrimSpace(string(cmdOutput))

	log.WithFields(log.Fields{
		"package":   packagePath,
		"cmdOutput": cmdOutput,
	}).Debug("installed package")

	return retval, nil
}
