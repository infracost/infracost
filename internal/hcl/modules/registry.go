package modules

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/infracost/infracost/internal/util"

	"github.com/hashicorp/go-retryablehttp"
	goversion "github.com/hashicorp/go-version"
	svchost "github.com/hashicorp/terraform-svchost"
	"github.com/hashicorp/terraform-svchost/auth"
	"github.com/hashicorp/terraform-svchost/disco"
	"github.com/rs/zerolog"

	"github.com/infracost/infracost/internal/apiclient"
	"github.com/infracost/infracost/internal/logging"
)

var defaultRegistryHost = "registry.terraform.io"

// validRegistryName is a regexp that matches valid registry identifier for namespaces, module names and targets
var validRegistryName = regexp.MustCompile("^[0-9A-Za-z-_]+$")

// The service ID used by Terraform Registry for discovering the module URLs
var moduleServiceID = "modules.v1"

// RegistryLookupResult is returned when looking up the module to check if it exists in the registry
// and has a matching version
type RegistryLookupResult struct {
	OK        bool
	ModuleURL RegistryURL
	Version   string
}

// RegistryURL contains given URL information for a module source. This can be used to build http requests to
// download the module or check versions of the module.
type RegistryURL struct {
	RawSource   string
	Host        string
	Location    string
	Credentials auth.HostCredentials
}

// Disco allows discovery on given hostnames. It tries to resolve a module source based on a set of
// discovery rules. It caches the results by hostname to avoid repeated requests for the same information.
// Therefore, it is advisable to use Disco per project and pass it to all required clients.
type Disco struct {
	disco           *disco.Disco
	logger          zerolog.Logger
	httpClient      *retryablehttp.Client
	proxyHttpClient *retryablehttp.Client

	locks sync.Map
}

// NewDisco returns a Disco with the provided credentialsSource initialising the underlying Terraform Disco.
// If Credentials are nil then all registry requests will be unauthed.
func NewDisco(credentialsSource auth.CredentialsSource, logger zerolog.Logger) *Disco {

	innerDisco := disco.NewWithCredentialsSource(credentialsSource)

	if proxyURL := os.Getenv("INFRACOST_REGISTRY_PROXY"); proxyURL != "" {
		innerDisco.Transport = &ConditionalTransport{
			ProxyHosts: []string{"terraform.io"},
			ProxyURL:   proxyURL,
			Inner:      innerDisco.Transport,
		}
	}

	return &Disco{
		disco:           innerDisco,
		logger:          logger,
		httpClient:      newRetryableClient(""),
		proxyHttpClient: newRetryableClient(os.Getenv("INFRACOST_REGISTRY_PROXY")),
	}
}

// ModuleLocation performs a discovery lookup for the given source and returns a RegistryURL with the real
// url of the module source and any required Credential information. It returns false if the module location
// is not recognised as a registry module.
func (d *Disco) ModuleLocation(source string) (RegistryURL, bool, error) {
	// So we expect them to only have 3 or 4 parts depending on if they explicitly specify the registry
	parts := strings.Split(source, "/")
	if len(parts) != 4 {
		return RegistryURL{}, false, fmt.Errorf("registry module source %s is not in the correct format", source)
	}

	host, namespace, moduleName, target := parts[0], parts[1], parts[2], parts[3]
	hostname, err := svchost.ForComparison(host)
	if err != nil {
		return RegistryURL{}, false, fmt.Errorf("unable to use user-provided module host %s as a Hostname for credential discovery: %w", host, err)
	}

	// lock the hostname to check credentials for as the underlying Terraform disco isn't concurrent safe and panics
	// with a concurrent map write error
	value, _ := d.locks.LoadOrStore(hostname, &sync.Mutex{})
	lock := value.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()

	serviceURL, err := d.disco.DiscoverServiceURL(hostname, moduleServiceID)
	if err != nil {
		return RegistryURL{}, false, fmt.Errorf("unable to discover registry service using host %s %w", host, err)
	}
	if !strings.HasSuffix(serviceURL.Path, "/") {
		serviceURL.Path += "/"
	}

	r := RegistryURL{
		Host:      host,
		Location:  fmt.Sprintf("%s%s/%s/%s", serviceURL.String(), namespace, moduleName, target),
		RawSource: source,
	}

	c, err := d.disco.CredentialsForHost(hostname)
	if err != nil {
		return r, true, fmt.Errorf("unable to retrieve credentials for registry host %s %w", host, err)
	}

	r.Credentials = c
	return r, true, nil
}

func (d *Disco) DownloadLocation(moduleURL RegistryURL, version string) (string, error) {
	hostname, err := svchost.ForComparison(moduleURL.Host)
	if err != nil {
		return "", fmt.Errorf("unable to use module URL %s as a Hostname to discover service URL: %w", moduleURL, err)
	}

	serviceURL, err := d.disco.DiscoverServiceURL(hostname, moduleServiceID)
	if err != nil {
		return "", fmt.Errorf("unable to discover registry service using host %s %w", moduleURL.Host, err)
	}
	if !strings.HasSuffix(serviceURL.Path, "/") {
		serviceURL.Path += "/"
	}

	var u *url.URL
	if version == "" {
		u, err = url.Parse(fmt.Sprintf("%s/download", moduleURL.Location))
	} else {
		u, err = url.Parse(fmt.Sprintf("%s/%s/download", moduleURL.Location, version))
	}
	if err != nil {
		return "", fmt.Errorf("error constructing download URL: %w", err)
	}

	downloadURL := serviceURL.ResolveReference(u)

	d.logger.Debug().Msgf("Looking up download URL for module %s from registry URL %s", moduleURL.RawSource, downloadURL.String())

	req, _ := http.NewRequest("GET", downloadURL.String(), nil)
	moduleURL.Credentials.PrepareRequest(req)
	retryReq, _ := retryablehttp.FromRequest(req)

	var client *retryablehttp.Client
	if moduleURL.Host == defaultRegistryHost || strings.HasSuffix(moduleURL.Host, ".terraform.io") {
		client = d.proxyHttpClient
	} else {
		client = d.httpClient
	}

	resp, err := client.Do(retryReq)

	if err != nil {
		return "", fmt.Errorf("error fetching download URL '%s': %w", util.RedactUrl(downloadURL.String()), err)
	}
	defer resp.Body.Close()

	location := resp.Header.Get("X-Terraform-Get")
	if location == "" {
		return "", fmt.Errorf("download URL has no X-Terraform-Get header, response status code: %d", resp.StatusCode)
	}

	if strings.HasPrefix(location, "/") || strings.HasPrefix(location, "./") || strings.HasPrefix(location, "../") {
		d.logger.Debug().Msgf("Detected relative path for location returned by download URL %s", downloadURL.String())

		locationURL, err := url.Parse(location)
		if err != nil {
			return "", fmt.Errorf("error parsing location URL: %w", err)
		}
		locationURL = serviceURL.ResolveReference(locationURL)
		location = locationURL.String()
	}

	return location, nil
}

func newRetryableClient(proxyURL string) *retryablehttp.Client {
	httpClient := retryablehttp.NewClient()
	httpClient.Logger = &apiclient.LeveledLogger{Logger: logging.Logger.With().Str("library", "retryablehttp").Logger()}
	if proxyURL != "" {
		if parsed, err := url.Parse(proxyURL); err == nil {
			httpClient.HTTPClient.Transport = &http.Transport{
				Proxy: http.ProxyURL(parsed),
			}
		}
	}
	return httpClient
}

// RegistryLoader is a loader that can lookup modules from a Terraform Registry and download them to the given destination
type RegistryLoader struct {
	packageFetcher  *PackageFetcher
	disco           *Disco
	logger          zerolog.Logger
	httpClient      *retryablehttp.Client
	proxyHttpClient *retryablehttp.Client
}

// NewRegistryLoader constructs a registry loader
func NewRegistryLoader(packageFetcher *PackageFetcher, disco *Disco, logger zerolog.Logger) *RegistryLoader {
	return &RegistryLoader{
		packageFetcher:  packageFetcher,
		disco:           disco,
		logger:          logger,
		httpClient:      newRetryableClient(""),
		proxyHttpClient: newRetryableClient(os.Getenv("INFRACOST_REGISTRY_PROXY")),
	}
}

// lookupModule lookups the matching version and download URL for the module.
// It calls the registry versions endpoint and tries to find a matching version.
func (r *RegistryLoader) lookupModule(moduleAddr string, versionConstraints string) (*RegistryLookupResult, error) {
	registrySource, err := normalizeRegistrySource(moduleAddr)
	if err != nil {
		r.logger.Debug().Err(err).Msgf("module '%s' not detected as registry module", util.RedactUrl(moduleAddr))
		return &RegistryLookupResult{
			OK: false,
		}, nil
	}

	moduleURL, ok, err := r.disco.ModuleLocation(registrySource)
	if !ok {
		if err != nil {
			r.logger.Debug().Err(err).Msgf("module '%s' not detected as registry module", util.RedactUrl(moduleAddr))
		}
		return &RegistryLookupResult{
			OK: false,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load remote module from given source %s and version constraints %s %w", util.RedactUrl(moduleAddr), versionConstraints, err)
	}

	versions, err := r.fetchModuleVersions(moduleURL)
	if err != nil {
		return nil, err
	}

	if len(versions) == 0 {
		return nil, errors.New("No versions found for registry module")
	}

	matchingVersion, err := findLatestMatchingVersion(versions, versionConstraints)
	if err != nil {
		return nil, err
	}

	return &RegistryLookupResult{
		OK:        true,
		ModuleURL: moduleURL,
		Version:   matchingVersion,
	}, nil
}

// fetchModuleVersions fetches the list of versions from the registry endpoint for the given module URL
func (r *RegistryLoader) fetchModuleVersions(moduleURL RegistryURL) ([]string, error) {
	req, _ := http.NewRequest("GET", moduleURL.Location+"/versions", nil)
	moduleURL.Credentials.PrepareRequest(req)
	retryReq, _ := retryablehttp.FromRequest(req)

	var client *retryablehttp.Client
	if moduleURL.Host == defaultRegistryHost || strings.HasSuffix(moduleURL.Host, ".terraform.io") {
		client = r.proxyHttpClient
	} else {
		client = r.httpClient
	}
	resp, err := client.Do(retryReq)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch registry module versions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Module versions endpoint returned status code %d", resp.StatusCode)
	}

	var versionsResp struct {
		Modules []struct {
			Versions []struct {
				Version string
			}
		}
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read module versions response: %w", err)
	}

	err = json.Unmarshal(respBody, &versionsResp)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal module versions response: %w", err)
	}

	if len(versionsResp.Modules) == 0 {
		return nil, fmt.Errorf("Module versions endpoint returned no modules")
	}

	versions := make([]string, 0, len(versionsResp.Modules[0].Versions))

	for _, v := range versionsResp.Modules[0].Versions {
		versions = append(versions, v.Version)
	}

	return versions, nil
}

// downloadModule downloads the module to the loader's destination
// It first calls the download URL to get the X-Terraform-Get header which contains a source we can use with go-getter to download the module
func (r *RegistryLoader) downloadModule(lookupResult *RegistryLookupResult, dest string) error {
	downloadURL, err := r.disco.DownloadLocation(lookupResult.ModuleURL, lookupResult.Version)
	if err != nil {
		return fmt.Errorf("could not find download location: %w", err)
	}
	// Deliberately not logging the download URL since it contains a token
	r.logger.Debug().Msgf("Downloading module %s", lookupResult.ModuleURL.RawSource)

	return r.packageFetcher.Fetch(downloadURL, dest)
}

func (r *RegistryLoader) DownloadLocation(moduleURL RegistryURL, version string) (string, error) {
	downloadURL, err := r.disco.DownloadLocation(moduleURL, version)
	if err != nil {
		return "", fmt.Errorf("could not find download location: %w", err)
	}

	return downloadURL, nil
}

// findLatestMatchingVersion returns the latest version from a list of versions that matches the given constraint.
// The constraints can be in any format that go-version understands, for example: "1.2.0", "~> 1.0", ">= 1.0, < 1.4"
// If the constraints are empty then the latest version is returned
// See https://www.terraform.io/language/expressions/version-constraints for more information on the version contraints
func findLatestMatchingVersion(versions []string, constraints string) (string, error) {
	// We now have a list of versions for the module, so we need to find the latest matching version
	var c goversion.Constraints
	var err error

	if constraints != "" {
		c, err = goversion.NewConstraint(constraints)
		if err != nil {
			return "", err
		}
	}

	var matchingVersion *goversion.Version

	// Loop through all the versions since they aren't necessarily sorted
	// Skip any versions that are less than the current matching version
	for _, rawVersion := range versions {
		version, err := goversion.NewVersion(rawVersion)
		if err != nil {
			return "", err
		}

		if matchingVersion != nil && version.LessThan(matchingVersion) {
			continue
		}

		// If there's no constraints then we want the latest version
		// Otherwise we need to check if the version matches the constraints
		if c.String() == "" || c.Check(version) {
			matchingVersion = version
		}
	}

	if matchingVersion == nil {
		return "", fmt.Errorf("No matching version found for constraint %s", constraints)
	}

	return matchingVersion.String(), nil
}

// normalizeRegistrySource validates a module source address and normalizes it into the host/namespace/module/target format
// This does not mean that the module address is a registry module, it could still be a remote module.
// To work that out we need to try looking up the module using the `lookupModule` function
func normalizeRegistrySource(moduleAddr string) (string, error) {
	// Modules are in the format (registry)/namspace/module/target
	// So we expect them to only have 3 or 4 parts depending on if they explicitly specify the registry
	parts := strings.Split(moduleAddr, "/")
	if len(parts) != 3 && len(parts) != 4 {
		return "", errors.New("Registry module source is not in the correct format")
	}

	// If the registry is not specified, we assume the default registry
	var host string
	var err error

	if len(parts) == 4 {
		host, err = normalizeHost(parts[0])
		if err != nil {
			return "", err
		}

		parts = parts[1:]
	} else {
		host = defaultRegistryHost
	}

	// GitHub and BitBucket hosts aren't supported as registries
	if host == "github.com" || host == "bitbucket.org" {
		return "", errors.New("Registry module source can not be from a GitHub or BitBucket host")
	}

	// Check that the other parts of the module source are using only the characters we expect
	namespace, moduleName, target := parts[0], parts[1], parts[2]
	if !validRegistryName.MatchString(namespace) || !validRegistryName.MatchString(moduleName) || !validRegistryName.MatchString(target) {
		return "", errors.New("Registry module source contains invalid characters")
	}

	return fmt.Sprintf("%s/%s/%s/%s", host, namespace, moduleName, target), nil
}

// normalizeHost extracts the hostname from the URL and normalizes it by:
// - Stripping the scheme (the leading "https://" or "http://")
// - Stripping anything trailing the hostname
// - Removing the port if it is the default 443 port
func normalizeHost(host string) (string, error) {
	var err error
	var parsedURL *url.URL

	parsedURL, err = url.Parse(host)
	if err != nil || parsedURL.Hostname() == "" {
		parsedURL, err = url.Parse("https://" + host)
		if err != nil || parsedURL.Hostname() == "" {
			return "", fmt.Errorf("Failed to parse host")
		}
	}

	portPart := ""

	port := parsedURL.Port()
	if port != "" && port != "443" {
		portPart = fmt.Sprintf(":%s", port)
	}

	return fmt.Sprintf("%s%s", parsedURL.Hostname(), portPart), nil
}
