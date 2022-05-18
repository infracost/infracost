package modules

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	goversion "github.com/hashicorp/go-version"
	svchost "github.com/hashicorp/terraform-svchost"
	"github.com/hashicorp/terraform-svchost/disco"
)

var defaultRegistryHost = "registry.terraform.io"

// validRegistryName is a regexp that matches valid registry identifier for namespaces, module names and targets
var validRegistryName = regexp.MustCompile("^[0-9A-Za-z-_]+$")

// RegistryLookupResult is returned when looking up the module to check if it exists in the registry
// and has a matching version
type RegistryLookupResult struct {
	Source      string
	Version     string
	DownloadURL string
	Token       string
}

// RegistryLoader is a loader that can lookup modules from a Terraform Registry and download them to the given destination
type RegistryLoader struct {
	packageFetcher    *PackageFetcher
	credentialsSource *CredentialsSource
}

// NewRegistryLoader constructs a registry loader
func NewRegistryLoader(packageFetcher *PackageFetcher, credentialsSource *CredentialsSource) *RegistryLoader {
	return &RegistryLoader{
		packageFetcher:    packageFetcher,
		credentialsSource: credentialsSource,
	}
}

// lookupModule lookups the matching version and download URL for the module.
// It calls the registry versions endpoint and tries to find a matching version.
func (r *RegistryLoader) lookupModule(moduleAddr string, versionConstraints string) (*RegistryLookupResult, error) {
	registrySource, err := normalizeRegistrySource(moduleAddr)
	if err != nil {
		return nil, err
	}

	// Modules are in the format (registry)/namspace/module/target
	// So we expect them to only have 3 or 4 parts depending on if they explicitly specify the registry
	parts := strings.Split(registrySource, "/")
	if len(parts) != 4 {
		return nil, errors.New("Registry module source is not in the correct format")
	}

	host, namespace, moduleName, target := parts[0], parts[1], parts[2], parts[3]

	services := disco.NewWithCredentialsSource(r.credentialsSource)
	serviceURL, err := services.DiscoverServiceURL(svchost.Hostname(host), "modules.v1")
	if err != nil {
		return nil, errors.New("Unable to discover registry services")
	}

	creds, err := services.CredentialsForHost(svchost.Hostname(host))
	if err != nil {
		return nil, errors.New("Unable to retrieve credentials for registry")
	}

	token := ""
	if creds != nil {
		token = creds.Token()
	}

	// By this stage we are more confident that the module source is a valid registry module
	// We now need to check the registry to see if the module exists and if it has a version
	moduleURL := fmt.Sprintf("%s%s/%s/%s", serviceURL.String(), namespace, moduleName, target)

	versions, err := r.fetchModuleVersions(moduleURL, token)
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
		Source:      fmt.Sprintf("%s%s/%s/%s", serviceURL.String(), namespace, moduleName, target),
		Version:     matchingVersion,
		DownloadURL: fmt.Sprintf("%s/%s/download", moduleURL, matchingVersion),
		Token:       token,
	}, nil
}

// fetchModuleVersions fetches the list of versions from the registry endpoint for the given module URL
func (r *RegistryLoader) fetchModuleVersions(moduleURL string, token string) ([]string, error) {
	httpClient := &http.Client{}
	req, _ := http.NewRequest("GET", moduleURL+"/versions", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := httpClient.Do(req)

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
func (r *RegistryLoader) downloadModule(downloadURL string, dest string, token string) error {
	httpClient := &http.Client{}
	req, _ := http.NewRequest("GET", downloadURL, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := httpClient.Do(req)

	if err != nil {
		return fmt.Errorf("Failed to download registry module: %w", err)
	}
	defer resp.Body.Close()

	source := resp.Header.Get("X-Terraform-Get")
	if source == "" {
		return errors.New("download URL has no X-Terraform-Get header")
	}

	return r.packageFetcher.fetch(source, dest)
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
