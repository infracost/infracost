package update

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/mod/semver"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/version"
)

type Info struct {
	LatestVersion string
	Cmd           string
}

func CheckForUpdate(ctx *config.RunContext) (*Info, error) {
	if skipUpdateCheck(ctx) {
		return nil, nil
	}

	// Check cache for the latest version
	cachedLatestVersion, err := checkCachedLatestVersion(ctx)
	if err != nil {
		logging.Logger.Debug().Msgf("error getting cached latest version: %v", err)
	}

	// Nothing to do if the current version is the latest cached version
	if cachedLatestVersion != "" && semver.Compare(version.Version, cachedLatestVersion) >= 0 {
		return nil, nil
	}

	isBrew, err := isBrewInstall()
	if err != nil {
		// don't fail if we can't detect brew, just fallback to other update method
		logging.Logger.Debug().Msgf("error checking if executable was installed via brew: %v", err)
	}

	var cmd string
	if isBrew {
		cmd = "$ brew upgrade infracost"
	} else {
		cmd = "Go to https://www.infracost.io/docs/update for instructions"
		if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
			cmd = "$ curl -fsSL https://raw.githubusercontent.com/infracost/infracost/master/scripts/install.sh | sh"
		}
	}

	// Get the latest version
	latestVersion := cachedLatestVersion
	if latestVersion == "" {
		if isBrew {
			latestVersion, err = getLatestBrewVersion()
		} else {
			latestVersion, err = getLatestGitHubVersion()
		}
		if err != nil {
			return nil, err
		}
	}

	// Save the latest version in the cache
	if latestVersion != cachedLatestVersion {
		err := setCachedLatestVersion(ctx, latestVersion)
		if err != nil {
			logging.Logger.Debug().Msgf("error saving cached latest version: %v", err)
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

func skipUpdateCheck(ctx *config.RunContext) bool {
	return ctx.Config.SkipUpdateCheck || config.IsTest() || config.IsDev()
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

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("Error getting latest version: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var parsedResp *formulaResp
	if json.Unmarshal(body, &parsedResp) != nil {
		return "", err
	}

	v := parsedResp.Versions.Stable
	if !strings.HasPrefix(v, "v") {
		v = fmt.Sprintf("v%s", v)
	}

	return v, nil
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var parsedResp *releaseResp
	if json.Unmarshal(body, &parsedResp) != nil {
		return "", err
	}

	v := parsedResp.TagName
	if !strings.HasPrefix(v, "v") {
		v = fmt.Sprintf("v%s", v)
	}

	return v, nil
}

func checkCachedLatestVersion(ctx *config.RunContext) (string, error) {
	if ctx.State.LatestReleaseCheckedAt == "" {
		return "", nil
	}

	checkedAt, err := time.Parse(time.RFC3339, ctx.State.LatestReleaseCheckedAt)
	if err != nil {
		return "", err
	}

	if checkedAt.Before(time.Now().Add(-24 * time.Hour)) {
		return "", nil
	}

	return ctx.State.LatestReleaseVersion, nil
}

func setCachedLatestVersion(ctx *config.RunContext, latestVersion string) error {
	ctx.State.LatestReleaseVersion = latestVersion
	ctx.State.LatestReleaseCheckedAt = time.Now().Format(time.RFC3339)

	return ctx.State.Save()
}
