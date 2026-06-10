package update

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
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

// installMethod identifies how the running binary was installed. The CLI
// has moved to github.com/infracost/cli for v2+, so the suggested upgrade
// command depends on which package manager (if any) owns the binary.
type installMethodKind int

const (
	installMethodUnknown installMethodKind = iota
	installMethodBrew
	installMethodChocolatey
)

// installMethod carries the detection result. ChocoPkg distinguishes the
// three Chocolatey packages we publish ("infracost", "infracost1",
// "infracost2") so the upgrade hint matches what choco actually expects.
type installMethod struct {
	Kind     installMethodKind
	ChocoPkg string
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

	method := detectInstallMethod()
	cmd := upgradeCommand(method)

	// Get the latest version
	latestVersion := cachedLatestVersion
	if latestVersion == "" {
		switch method.Kind {
		case installMethodBrew:
			latestVersion, err = getLatestBrewVersion()
		case installMethodChocolatey:
			latestVersion, err = getLatestChocolateyVersion(method.ChocoPkg)
		default:
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
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
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

	// Point at infracost/cli — the v2+ repo. This is the whole point of the
	// v0.10 patch: surface v2 to users still running the legacy series.
	resp, err := http.Get("https://api.github.com/repos/infracost/cli/releases/latest")
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

// detectInstallMethod is overridable in tests. Detection is best-effort —
// any error is swallowed and reported as unknown so a transient detection
// hiccup never blocks the user's command.
var detectInstallMethod = func() installMethod {
	if isBrew, err := isBrewInstall(); err == nil && isBrew {
		return installMethod{Kind: installMethodBrew}
	} else if err != nil {
		logging.Logger.Debug().Msgf("error checking brew install: %v", err)
	}

	if pkg := chocolateyPackage(); pkg != "" {
		return installMethod{Kind: installMethodChocolatey, ChocoPkg: pkg}
	}

	return installMethod{Kind: installMethodUnknown}
}

// chocolateyPackage returns the choco package that owns the running binary
// ("infracost", "infracost1", "infracost2"), or "" if the binary doesn't
// live under the Chocolatey install root. Detection relies on os.Executable
// resolving to <ChocolateyInstall>\lib\<pkg>\... — which is what happens
// when chocolatey's shim launches the real binary.
func chocolateyPackage() string {
	if runtime.GOOS != "windows" {
		return ""
	}

	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return ""
	}

	root := os.Getenv("ChocolateyInstall")
	if root == "" {
		root = `C:\ProgramData\chocolatey`
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return ""
	}

	return chocolateyPackageFromPath(exe, root)
}

// chocolateyPackageFromPath is the pure-function half of choco detection so
// it's testable without faking os.Executable() on non-Windows hosts. Both
// inputs are compared case-insensitively because Windows filesystems are
// case-insensitive in practice.
func chocolateyPackageFromPath(exe, root string) string {
	libPrefix := strings.ToLower(filepath.Join(root, "lib")) + string(filepath.Separator)
	exeLower := strings.ToLower(exe)
	if !strings.HasPrefix(exeLower, libPrefix) {
		return ""
	}

	rel := exeLower[len(libPrefix):]
	if i := strings.IndexRune(rel, filepath.Separator); i >= 0 {
		rel = rel[:i]
	}
	return rel
}

// upgradeCommand returns the shell hint we render in the update notice. For
// users on the legacy `infracost` choco package we also point out that
// installing `infracost1` keeps them on the v0.10 series, in case they
// don't want to roll forward to v2.
func upgradeCommand(method installMethod) string {
	switch method.Kind {
	case installMethodBrew:
		return "$ brew upgrade infracost"
	case installMethodChocolatey:
		pkg := method.ChocoPkg
		if pkg == "" {
			pkg = "infracost"
		}
		cmd := fmt.Sprintf("$ choco upgrade %s", pkg)
		if pkg == "infracost" {
			cmd += "\n(To stay on the v0.10 series, install the infracost1 package instead.)"
		}
		return cmd
	default:
		if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
			return "$ curl -fsSL https://raw.githubusercontent.com/infracost/cli/master/scripts/install.sh | sh"
		}
		return "Go to https://www.infracost.io/docs/update for instructions"
	}
}

// getLatestChocolateyVersion queries the community Chocolatey OData feed
// for the latest published version of the given package.
func getLatestChocolateyVersion(pkg string) (string, error) {
	if pkg == "" {
		pkg = "infracost"
	}

	resp, err := http.Get(chocolateyFeedURL(pkg))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("error getting latest chocolatey version: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	v, err := parseChocolateyFeed(body)
	if err != nil {
		return "", err
	}
	if v == "" {
		return "", errors.Errorf("chocolatey feed returned no version for %s", pkg)
	}
	return v, nil
}

// chocolateyFeedURL builds the OData filter URL. Package names are
// constrained to [a-z0-9.-] by Chocolatey, so no escaping is needed.
func chocolateyFeedURL(pkg string) string {
	return "https://community.chocolatey.org/api/v2/Packages()?$filter=Id%20eq%20%27" + pkg + "%27%20and%20IsLatestVersion"
}

// parseChocolateyFeed extracts the first non-empty <Version> from the ATOM
// feed body, prefixed with "v" for downstream semver comparison. Returns
// "" if the feed has no usable entry (caller decides how to report that).
func parseChocolateyFeed(body []byte) (string, error) {
	var feed struct {
		Entries []struct {
			Properties struct {
				Version string `xml:"Version"`
			} `xml:"properties"`
		} `xml:"entry"`
	}
	if err := xml.Unmarshal(body, &feed); err != nil {
		return "", errors.Wrap(err, "parsing chocolatey feed")
	}

	for _, e := range feed.Entries {
		v := strings.TrimSpace(e.Properties.Version)
		if v == "" {
			continue
		}
		if !strings.HasPrefix(v, "v") {
			v = "v" + v
		}
		return v, nil
	}
	return "", nil
}
