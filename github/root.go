package github

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type APIReleaseResponse []struct {
	URL       string `json:"url"`
	AssetsURL string `json:"assets_url"`
	UploadURL string `json:"upload_url"`
	HTMLURL   string `json:"html_url"`
	ID        int    `json:"id"`
	Author    struct {
		Login             string `json:"login"`
		ID                int    `json:"id"`
		NodeID            string `json:"node_id"`
		AvatarURL         string `json:"avatar_url"`
		GravatarID        string `json:"gravatar_id"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		OrganizationsURL  string `json:"organizations_url"`
		ReposURL          string `json:"repos_url"`
		EventsURL         string `json:"events_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		Type              string `json:"type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"author"`
	NodeID          string    `json:"node_id"`
	TagName         string    `json:"tag_name"`
	TargetCommitish string    `json:"target_commitish"`
	Name            string    `json:"name"`
	Draft           bool      `json:"draft"`
	Prerelease      bool      `json:"prerelease"`
	CreatedAt       time.Time `json:"created_at"`
	PublishedAt     time.Time `json:"published_at"`
	Assets          []struct {
		URL      string `json:"url"`
		ID       int    `json:"id"`
		NodeID   string `json:"node_id"`
		Name     string `json:"name"`
		Label    string `json:"label"`
		Uploader struct {
			Login             string `json:"login"`
			ID                int    `json:"id"`
			NodeID            string `json:"node_id"`
			AvatarURL         string `json:"avatar_url"`
			GravatarID        string `json:"gravatar_id"`
			URL               string `json:"url"`
			HTMLURL           string `json:"html_url"`
			FollowersURL      string `json:"followers_url"`
			FollowingURL      string `json:"following_url"`
			GistsURL          string `json:"gists_url"`
			StarredURL        string `json:"starred_url"`
			SubscriptionsURL  string `json:"subscriptions_url"`
			OrganizationsURL  string `json:"organizations_url"`
			ReposURL          string `json:"repos_url"`
			EventsURL         string `json:"events_url"`
			ReceivedEventsURL string `json:"received_events_url"`
			Type              string `json:"type"`
			SiteAdmin         bool   `json:"site_admin"`
		} `json:"uploader"`
		ContentType        string    `json:"content_type"`
		State              string    `json:"state"`
		Size               int       `json:"size"`
		DownloadCount      int       `json:"download_count"`
		CreatedAt          time.Time `json:"created_at"`
		UpdatedAt          time.Time `json:"updated_at"`
		BrowserDownloadURL string    `json:"browser_download_url"`
	} `json:"assets"`
	TarballURL string `json:"tarball_url"`
	ZipballURL string `json:"zipball_url"`
	Body       string `json:"body"`
}

type Release struct {
	Version     string    `json:"version"`
	Name        string    `json:"name"`
	DownloadUrl string    `json:"downloadUrl"`
	Created     time.Time `json:"createdDate"`
}

// GetVersionsForRepo gets the latest available assets for the given github repo (and all other versions?)
func GetVersionsForRepo(name, repo string) ([]Release, error) {
	retval := []Release{}
	releaseResponse := APIReleaseResponse{}

	//	Format our url:
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", name, repo)

	//	Create a request with headers
	clientRequest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.WithError(err).Error("problem preparing the request to the github api")
		return retval, err
	}

	//	Set our headers
	clientRequest.Header.Set("Content-Type", "application/json; charset=UTF-8")

	//	Execute the request
	client := &http.Client{}
	clientResponse, err := client.Do(clientRequest)
	if err != nil {
		log.WithError(err).Error("problem sending the request to the github api")
		return retval, err
	}
	defer clientResponse.Body.Close()

	//	Decode the response:
	err = json.NewDecoder(clientResponse.Body).Decode(&releaseResponse)
	if err != nil {
		log.WithError(err).Error("problem decoding the response from the github api")
		return retval, nil
	}

	//	Loop through each release
	for _, item := range releaseResponse {

		//	Analyze the assets.  If we have a .deb file, track it
		//	Don't append to the results if we don't have a .deb file
		for _, asset := range item.Assets {
			if strings.HasSuffix(asset.Name, ".deb") {

				//	Create a new release object with the release version and create date
				newRelease := Release{
					Version: item.TagName,
					Created: asset.CreatedAt,
				}

				//	Set the name and the url information
				newRelease.Name = asset.Name
				newRelease.DownloadUrl = asset.BrowserDownloadURL

				//	Add this release to the list
				retval = append(retval, newRelease)
			}
		}
	}

	return retval, nil
}

// DownloadFile downloads a remote file to a temporary location and returns the temporary location
func DownloadFile(remoteUrl string) (string, error) {

	//	Get a temporary file reference:
	tempPathLocation, err := ioutil.TempFile("appupgrade", "*.deb")
	if err != nil {
		log.WithError(err).Error("problem creating temp file")
		return "", err
	}

	//	Download the remote url to the temp file:
	resp, err := http.Get(remoteUrl)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"remoteUrl": remoteUrl,
		}).Error("problem downloading remote file")
		return "", err
	}
	defer resp.Body.Close()

	//	Save the downloaded file to the temp file:
	_, err = io.Copy(tempPathLocation, resp.Body)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"remoteUrl": remoteUrl,
		}).Error("problem saving remote file")
		return "", err
	}

	//	Return the local file path that contains the remote url contents
	return tempPathLocation.Name(), nil
}
