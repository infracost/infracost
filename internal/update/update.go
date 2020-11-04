package update

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/mod/semver"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/version"
)

type Info struct {
	LatestVersion string
	Cmd           string
}

func CheckForUpdate() (*Info, error) {
	if skipUpdateCheck() {
		return nil, nil
	}

	var cmd string

	state, err := config.ReadStateFileIfNotExists()
	if err != nil {
		log.Debugf("error reading state file: %v", err)
	}

	cachedRelease, err := checkCachedLatestVersion(state)
	if err != nil {
		log.Debugf("error getting cached latest version: %v", err)
	}

	latestVersion := cachedRelease

	isBrew, err := isBrewInstall()
	if err != nil {
		// don't fail if we can't detect brew, just fallback to other update method
		log.Debugf("error checking if executable was installed via brew: %v", err)
	}

	if isBrew && err == nil {
		if latestVersion == "" {
			latestVersion, err = getLatestBrewVersion()
			if err != nil {
				return nil, err
			}
		}

		cmd = "$ brew upgrade infracost"
	} else {
		if latestVersion == "" {
			latestVersion, err = getLatestGitHubVersion()
			if err != nil {
				return nil, err
			}
		}

		cmd = "Go to https://www.infracost.io/docs/update for instructions"
		if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
			cmd = "$ curl -s -L https://github.com/infracost/infracost/releases/latest/download/infracost-linux-amd64.tar.gz | tar xz -C /tmp && \\\n  sudo mv /tmp/infracost-linux-amd64 /usr/local/bin/infracost"
		} else if runtime.GOOS == "darwin" && runtime.GOARCH == "amd64" {
			cmd = "$ curl -s -L https://github.com/infracost/infracost/releases/latest/download/infracost-darwin-amd64.tar.gz | tar xz -C /tmp && \\\n  sudo mv /tmp/infracost-darwin-amd64 /usr/local/bin/infracost"
		}
	}

	if cachedRelease == "" {
		err := saveCachedLatestVersion(state, latestVersion)
		if err != nil {
			log.Debugf("error saving cached latest version: %v", err)
		}
	}

	if semver.Compare(version.Version, latestVersion) >= 0 {
		return nil, nil
	}

	return &Info{
		LatestVersion: latestVersion,
		Cmd:           cmd,
	}, nil
}

func skipUpdateCheck() bool {
	return config.IsTruthy(os.Getenv("INFRACOST_SKIP_UPDATE_CHECK")) || config.IsTest() || config.IsDev()
}

func isBrewInstall() (bool, error) {
	if runtime.GOOS != "darwin" {
		return false, nil
	}

	exe, err := os.Executable()
	if err != nil {
		return false, errors.Wrap(err, "error finding infracost executable")
	}

	path, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return false, errors.Wrap(err, "error evaluating infracost executable symlink")
	}

	brewPrefixCmd := exec.Command("brew", "--prefix", "infracost")

	var stdout bytes.Buffer
	brewPrefixCmd.Stdout = &stdout

	err = brewPrefixCmd.Run()
	if err != nil {
		return false, errors.Wrap(err, "error running 'brew --prefix infracost'")
	}

	brewPrefixPath, err := filepath.EvalSymlinks(strings.TrimSpace(stdout.String()))
	if err != nil {
		return false, errors.Wrap(err, "error evaluating brew prefix path symlink")
	}

	brewInfracostPath := filepath.Join(brewPrefixPath, "bin", "infracost")

	return path == brewInfracostPath, nil
}

func getLatestBrewVersion() (string, error) {
	type versionsResp struct {
		Stable string `json:"stable"`
	}

	type formulaResp struct {
		Versions versionsResp `json:"versions"`
	}

	resp, err := http.Get("https://formulae.brew.sh/api/formula/infracost.json")
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var parsedResp *formulaResp
	if json.Unmarshal(body, &parsedResp) != nil {
		return "", err
	}

	return parsedResp.Versions.Stable, nil
}

func getLatestGitHubVersion() (string, error) {
	type releaseResp struct {
		TagName string `json:"tag_name"`
	}

	resp, err := http.Get("https://api.github.com/repos/infracost/infracost/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var parsedResp *releaseResp
	if json.Unmarshal(body, &parsedResp) != nil {
		return "", err
	}

	return parsedResp.TagName, nil
}

func checkCachedLatestVersion(state config.StateFile) (string, error) {
	if state.LatestReleaseCheckedAt == "" {
		return "", nil
	}

	checkedAt, err := time.Parse(time.RFC3339, state.LatestReleaseCheckedAt)
	if err != nil {
		return "", err
	}

	if checkedAt.Before(time.Now().Add(-24 * time.Hour)) {
		return "", nil
	}

	return state.LatestReleaseVersion, nil
}

func saveCachedLatestVersion(state config.StateFile, latestVersion string) error {
	state.LatestReleaseVersion = latestVersion
	state.LatestReleaseCheckedAt = time.Now().Format(time.RFC3339)

	return config.WriteStateFile(state)
}
