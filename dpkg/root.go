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
