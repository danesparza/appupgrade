package github_test

import (
	"testing"

	"github.com/danesparza/appupgrade/github"
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
