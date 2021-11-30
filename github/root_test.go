package github_test

import (
	"testing"

	"github.com/danesparza/appupgrade/github"
	"github.com/hashicorp/go-version"
)

func TestGithub_GetVersionsForRepo_ValidRepo_Successful(t *testing.T) {

	//	Arrange
	user := "danesparza"
	repo := "daydash"

	//	Act
	releases, err := github.GetVersionsForRepo(user, repo)

	//	Assert
	if err != nil {
		t.Errorf("GetVersionsForRepo - Should get versions without error, but got: %s", err)
	}

	if len(releases) < 1 {
		t.Errorf("GetVersionsForRepo failed: Should have fetched releases, but got none")
	}

	t.Logf("%+v", releases)

}

func TestGithub_CompareVersions_ValidRepo_Successful(t *testing.T) {

	//	Arrange
	user := "danesparza"
	repo := "daydash"
	compareVersion := "v1.0.0"

	//	Act
	releases, err := github.GetVersionsForRepo(user, repo)

	//	Assert
	if err != nil {
		t.Errorf("GetVersionsForRepo - Should get versions without error, but got: %s", err)
	}

	if len(releases) < 1 {
		t.Errorf("GetVersionsForRepo failed: Should have fetched releases, but got none")
	}

	//	Compare versions:
	v1, err := version.NewVersion(compareVersion)
	if err != nil {
		t.Errorf("NewVersion - failed to parse compareVersion: %s", err)
	}

	v2, err := version.NewVersion(releases[0].Version)
	if err != nil {
		t.Errorf("NewVersion - failed to parse releases[0].Version: %s", err)
	}

	if v1.LessThan(v2) {
		t.Logf("%s is less than %s", v1, v2)
	} else {
		t.Logf("%s is greater than %s", v1, v2)
	}
}
