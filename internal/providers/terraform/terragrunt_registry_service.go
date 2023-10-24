package terraform

// Contents of this file were copied from github.com/gruntwork-io/terragrunt
import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-getter"
	safetemp "github.com/hashicorp/go-safetemp"

	"github.com/gruntwork-io/go-commons/errors"
	"github.com/gruntwork-io/terragrunt/util"
)

// httpClient is the default client to be used by HttpGetters.
var httpClient = cleanhttp.DefaultClient()

// Constants relevant to the module registry
const (
	defaultRegistryDomain = "registry.terraform.io"
	serviceDiscoveryPath  = "/.well-known/terraform.json"
	versionQueryKey       = "version"
	authTokenEnvVarName   = "TG_TF_REGISTRY_TOKEN" //nolint
)

// TerraformRegistryServicePath is a struct for extracting the modules service path in the Registry.
type TerraformRegistryServicePath struct {
	ModulesPath string `json:"modules.v1"`
}

// TerraformRegistryGetter is a Getter (from go-getter) implementation that will download from the terraform module
// registry. This supports getter URLs encoded in the following manner:
//
// tfr://REGISTRY_DOMAIN/MODULE_PATH?version=VERSION
//
// Where the REGISTRY_DOMAIN is the terraform registry endpoint (e.g., registry.terraform.io), MODULE_PATH is the
// registry path for the module (e.g., terraform-aws-modules/vpc/aws), and VERSION is the specific version of the module
// to download (e.g., 2.2.0).
//
// This protocol will use the Module Registry Protocol (documented at
// https://www.terraform.io/docs/internals/module-registry-protocol.html) to lookup the module source URL and download
// it.
//
// Authentication to private module registries is handled via environment variables. The authorization API token is
// expected to be provided to Terragrunt via the TG_TF_REGISTRY_TOKEN environment variable. This token can be any
// registry API token generated on Terraform Cloud / Enterprise.
//
// MAINTAINER'S NOTE: Ideally we implement the full credential system that terraform uses as part of `terraform login`,
// but all the relevant packages are internal to the terraform repository, thus making it difficult to use as a
// library. For now, we keep things simple by supporting providing tokens via env vars and in the future, we can
// consider implementing functionality to load credentials from terraform.
// GH issue: https://github.com/gruntwork-io/terragrunt/issues/1771
//
// MAINTAINER'S NOTE: Ideally we can support a shorthand notation that omits the tfr:// protocol to detect that it is
// referring to a terraform registry, but this requires implementing a complex detector and ensuring it has precedence
// over the file detector. We deferred the implementation for that to a future release.
// GH issue: https://github.com/gruntwork-io/terragrunt/issues/1772
type TerraformRegistryGetter struct {
	client *getter.Client
}

// SetClient allows the getter to know what getter client (different from the underlying HTTP client) to use for
// progress tracking.
func (tfrGetter *TerraformRegistryGetter) SetClient(client *getter.Client) {
	tfrGetter.client = client
}

// Context returns the go context to use for the underlying fetch routines. This depends on what client is set.
func (tfrGetter *TerraformRegistryGetter) Context() context.Context {
	if tfrGetter == nil || tfrGetter.client == nil {
		return context.Background()
	}
	return tfrGetter.client.Ctx
}

// ClientMode returns the download mode based on the given URL. Since this getter is designed around the Terraform
// module registry, we always use Dir mode so that we can download the full Terraform module.
func (tfrGetter *TerraformRegistryGetter) ClientMode(u *url.URL) (getter.ClientMode, error) {
	return getter.ClientModeDir, nil
}

// Get is the main routine to fetch the module contents specified at the given URL and download it to the dstPath.
// This routine assumes that the srcURL points to the Terraform registry URL, with the Path configured to the module
// path encoded as `:namespace/:name/:system` as expected by the Terraform registry. Note that the URL query parameter
// must have the `version` key to specify what version to download.
func (tfrGetter *TerraformRegistryGetter) Get(dstPath string, srcURL *url.URL) error {
	ctx := tfrGetter.Context()

	registryDomain := srcURL.Host
	if registryDomain == "" {
		registryDomain = defaultRegistryDomain
	}
	queryValues := srcURL.Query()
	modulePath, moduleSubDir := getter.SourceDirSubdir(srcURL.Path)

	versionList, hasVersion := queryValues[versionQueryKey]
	if !hasVersion {
		return errors.WithStackTrace(MalformedRegistryURLErr{reason: "missing version query"})
	}
	if len(versionList) != 1 {
		return errors.WithStackTrace(MalformedRegistryURLErr{reason: "more than one version query"})
	}
	version := versionList[0]

	moduleRegistryBasePath, err := getModuleRegistryURLBasePath(ctx, registryDomain)
	if err != nil {
		return err
	}

	moduleURL, err := buildRequestUrl(registryDomain, moduleRegistryBasePath, modulePath, version)
	if err != nil {
		return err
	}

	terraformGet, err := getTerraformGetHeader(ctx, *moduleURL)
	if err != nil {
		return err
	}

	downloadURL, err := getDownloadURLFromHeader(*moduleURL, terraformGet)
	if err != nil {
		return err
	}

	// If there is a subdir component, then we download the root separately into a temporary directory, then copy over
	// the proper subdir. Note that we also have to take into account sub dirs in the original URL in addition to the
	// subdir component in the X-Terraform-Get download URL.
	source, subDir := getter.SourceDirSubdir(downloadURL)
	if subDir == "" && moduleSubDir == "" {
		var opts []getter.ClientOption
		if tfrGetter.client != nil {
			opts = tfrGetter.client.Options
		}
		return getter.Get(dstPath, source, opts...)
	}

	// We have a subdir, time to jump some hoops
	return tfrGetter.getSubdir(ctx, dstPath, source, path.Join(subDir, moduleSubDir))
}

// GetFile is not implemented for the Terraform module registry Getter since the terraform module registry doesn't serve
// a single file.
func (tfrGetter *TerraformRegistryGetter) GetFile(dst string, src *url.URL) error {
	return errors.WithStackTrace(fmt.Errorf("GetFile is not implemented for the Terraform Registry Getter"))
}

// getSubdir downloads the source into the destination, but with the proper subdir.
func (tfrGetter *TerraformRegistryGetter) getSubdir(ctx context.Context, dstPath, sourceURL, subDir string) error {
	// Create a temporary directory to store the full source. This has to be a non-existent directory.
	tempdirPath, tempdirCloser, err := safetemp.Dir("", "getter")
	if err != nil {
		return err
	}
	defer tempdirCloser.Close()

	var opts []getter.ClientOption
	if tfrGetter.client != nil {
		opts = tfrGetter.client.Options
	}
	// Download that into the given directory
	if err := getter.Get(tempdirPath, sourceURL, opts...); err != nil {
		return errors.WithStackTrace(err)
	}

	// Process any globbing
	sourcePath, err := getter.SubdirGlob(tempdirPath, subDir)
	if err != nil {
		return errors.WithStackTrace(err)
	}

	// Make sure the subdir path actually exists
	if _, err := os.Stat(sourcePath); err != nil {
		details := fmt.Sprintf("could not stat download path %s (error: %s)", sourcePath, err)
		return errors.WithStackTrace(ModuleDownloadErr{sourceURL: sourceURL, details: details})
	}

	// Copy the subdirectory into our actual destination.
	if err := os.RemoveAll(dstPath); err != nil {
		return errors.WithStackTrace(err)
	}

	// Make the final destination
	if err := os.MkdirAll(dstPath, 0755); err != nil {
		return errors.WithStackTrace(err)
	}

	// We use a temporary manifest file here that is deleted at the end of this routine since we don't intend to come
	// back to it.
	manifestFname := ".tgmanifest"
	manifestPath := filepath.Join(dstPath, manifestFname)
	defer os.Remove(manifestPath)
	return util.CopyFolderContentsWithFilter(sourcePath, dstPath, manifestFname, func(path string) bool { return true })
}

// getModuleRegistryURLBasePath uses the service discovery protocol
// (https://www.terraform.io/docs/internals/remote-service-discovery.html)
// to figure out where the modules are stored. This will return the base
// path where the modules can be accessed
func getModuleRegistryURLBasePath(ctx context.Context, domain string) (string, error) {
	sdURL := url.URL{
		Scheme: "https",
		Host:   domain,
		Path:   serviceDiscoveryPath,
	}
	bodyData, _, err := httpGETAndGetResponse(ctx, sdURL)
	if err != nil {
		return "", err
	}

	var respJSON TerraformRegistryServicePath
	if err := json.Unmarshal(bodyData, &respJSON); err != nil {
		reason := fmt.Sprintf("Error parsing response body %s: %s", string(bodyData), err)
		return "", errors.WithStackTrace(ServiceDiscoveryErr{reason: reason})
	}
	return respJSON.ModulesPath, nil
}

// getTerraformGetHeader makes an http GET call to the given registry URL and return the contents of the header
// X-Terraform-Get. This function will return an error if the response does not contain the header.
func getTerraformGetHeader(ctx context.Context, url url.URL) (string, error) {
	_, header, err := httpGETAndGetResponse(ctx, url)
	if err != nil {
		details := "error receiving HTTP data"
		return "", errors.WithStackTrace(ModuleDownloadErr{sourceURL: url.String(), details: details})
	}

	terraformGet := header.Get("X-Terraform-Get")
	if terraformGet == "" {
		details := "no source URL was returned in header X-Terraform-Get from download URL"
		return "", errors.WithStackTrace(ModuleDownloadErr{sourceURL: url.String(), details: details})
	}
	return terraformGet, nil
}

// getDownloadURLFromHeader checks if the content of the X-Terraform-GET header contains the base url
// and prepends it if not
func getDownloadURLFromHeader(moduleURL url.URL, terraformGet string) (string, error) {
	// If url from X-Terrafrom-Get Header seems to be a relative url,
	// append scheme and host from url used for getting the download url
	// because third-party registry implementations may not "know" their own absolute URLs if
	// e.g. they are running behind a reverse proxy frontend, or such.
	if strings.HasPrefix(terraformGet, "/") || strings.HasPrefix(terraformGet, "./") || strings.HasPrefix(terraformGet, "../") {
		relativePathURL, err := url.Parse(terraformGet)
		if err != nil {
			return "", errors.WithStackTrace(err)
		}
		terraformGetURL := moduleURL.ResolveReference(relativePathURL)
		terraformGet = terraformGetURL.String()
	}
	return terraformGet, nil
}

// httpGETAndGetResponse is a helper function to make a GET request to the given URL using the http client. This
// function will then read the response and return the contents + the response header.
func httpGETAndGetResponse(ctx context.Context, getURL url.URL) ([]byte, *http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", getURL.String(), nil)
	if err != nil {
		return nil, nil, errors.WithStackTrace(err)
	}

	// Handle authentication via env var. Authentication is done by providing the registry token as a bearer token in
	// the request header.
	authToken := os.Getenv(authTokenEnvVarName)
	if authToken != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, nil, errors.WithStackTrace(err)
	}

	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, nil, errors.WithStackTrace(RegistryAPIErr{url: getURL.String(), statusCode: resp.StatusCode})
	}

	bodyData, err := io.ReadAll(resp.Body)
	return bodyData, &resp.Header, errors.WithStackTrace(err)
}

// buildRequestUrl - create url to download module using moduleRegistryBasePath
func buildRequestUrl(registryDomain string, moduleRegistryBasePath string, modulePath string, version string) (*url.URL, error) {
	moduleRegistryBasePath = strings.TrimSuffix(moduleRegistryBasePath, "/")
	modulePath = strings.TrimSuffix(modulePath, "/")
	modulePath = strings.TrimPrefix(modulePath, "/")

	moduleFullPath := fmt.Sprintf("%s/%s/%s/download", moduleRegistryBasePath, modulePath, version)

	moduleURL, err := url.Parse(moduleFullPath)
	if err != nil {
		return nil, err
	}
	if moduleURL.Scheme != "" {
		return moduleURL, nil
	}
	return &url.URL{Scheme: "https", Host: registryDomain, Path: moduleFullPath}, nil
}

// MalformedRegistryURLErr is returned if the Terraform Registry URL passed to the Getter is malformed.
type MalformedRegistryURLErr struct {
	reason string
}

func (err MalformedRegistryURLErr) Error() string {
	return fmt.Sprintf("tfr getter URL is malformed: %s", err.reason)
}

// ServiceDiscoveryErr is returned if Terragrunt failed to identify the module API endpoint through the service
// discovery protocol.
type ServiceDiscoveryErr struct {
	reason string
}

func (err ServiceDiscoveryErr) Error() string {
	return fmt.Sprintf("Error identifying module registry API location: %s", err.reason)
}

// ModuleDownloadErr is returned if Terragrunt failed to download the module.
type ModuleDownloadErr struct {
	sourceURL string
	details   string
}

func (err ModuleDownloadErr) Error() string {
	return fmt.Sprintf("Error downloading module from %s: %s", err.sourceURL, err.details)
}

// RegistryAPIErr is returned if we get an unsuccessful HTTP return code from the registry.
type RegistryAPIErr struct {
	url        string
	statusCode int
}

func (err RegistryAPIErr) Error() string {
	return fmt.Sprintf("Failed to fetch url %s: status code %d", err.url, err.statusCode)
}
