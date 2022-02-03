package hcl

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"regexp"
	"strings"

	getter "github.com/hashicorp/go-getter"
	goversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/pkg/errors"
)

var infracostDir = ".infracost"
var modulesCacheDir = path.Join(infracostDir, "modules")

type ModuleMetadata struct {
	Key     string `json:"Key"`
	Source  string `json:"Source"`
	Version string `json:"Version"`
	Dir     string `json:"Dir"`
}

type RegistryModuleCheckResult struct {
	Version     string
	DownloadURL string
}

type ModuleLoader struct {
	path string
}

var validRegistryName = regexp.MustCompile("^[0-9A-Za-z-_]+$")

func (m *ModuleLoader) Load() ([]ModuleMetadata, error) {
	metadatas := make([]ModuleMetadata, 0)

	module, diags := tfconfig.LoadModule(m.path)
	if diags.HasErrors() {
		return metadatas, diags.Err()
	}

	for _, moduleCall := range module.ModuleCalls {
		metadata, err := m.loadModule(moduleCall)
		if err != nil {
			return metadatas, err
		}

		metadatas = append(metadatas, metadata)
	}

	return metadatas, nil
}

func (m *ModuleLoader) loadModule(moduleCall *tfconfig.ModuleCall) (ModuleMetadata, error) {
	metadata := ModuleMetadata{
		Key:    moduleCall.Name,
		Source: moduleCall.Source,
	}

	if m.isLocalModule(moduleCall) {
		metadata.Dir = moduleCall.Source
		return metadata, nil
	}

	if checkResult, err := m.checkRegistryModule(moduleCall); err == nil {
		// TODO: figure out a better path for this to support:
		// - nested modules with the same name
		// - modules with the same source not needing downloaded twice
		dir := path.Join(modulesCacheDir, moduleCall.Name)

		err := m.downloadRegistryModule(checkResult.DownloadURL, dir)
		if err != nil {
			return metadata, err
		}

		metadata.Version = checkResult.Version
		metadata.Dir = dir
		return metadata, nil
	}

	// TODO: figure out a better path for this to support:
	// - nested modules with the same name
	// - modules with the same source not needing downloaded twice
	dir := path.Join(modulesCacheDir, moduleCall.Name)

	err := m.downloadRemoteModule(moduleCall.Source, dir)
	if err != nil {
		return metadata, err
	}

	metadata.Dir = dir
	return metadata, nil
}

func (m *ModuleLoader) isLocalModule(moduleCall *tfconfig.ModuleCall) bool {
	return strings.HasPrefix(moduleCall.Source, ".")
}

func (m *ModuleLoader) checkRegistryModule(moduleCall *tfconfig.ModuleCall) (*RegistryModuleCheckResult, error) {
	// Modules are in the format (registry)/namspace/module/target
	// So we expect them to only have 3 or 4 parts depending on if they explicitly specify the registry
	parts := strings.Split(moduleCall.Source, "/")
	if len(parts) != 3 && len(parts) != 4 {
		return nil, errors.New("Registry module source is not in the correct format")
	}

	// If the registry is not specified, we assume the default registry
	var host string
	if len(parts) == 4 {
		host = parts[0]
		parts = parts[1:]
	} else {
		host = "registry.terraform.io"
	}

	// GitHub and BitBucket hosts aren't supported as registries
	// TODO: check "friendly" hostname
	if host == "github.com" || host == "bitbucket.org" {
		return nil, errors.New("Registry module source can not be from a GitHub or BitBucket host")
	}

	// Check that the other parts of the module source are using only the characters we expect
	namespace, moduleName, target := parts[0], parts[1], parts[2]
	if !validRegistryName.MatchString(namespace) || !validRegistryName.MatchString(moduleName) || !validRegistryName.MatchString(target) {
		return nil, errors.New("Registry module source contains invalid characters")
	}

	// By this stage we are more confident that the module source is a valid registry module
	// We now need to check the registry to see if the module exists and if it has a version
	moduleURL := fmt.Sprintf("https://%s/v1/modules/%s/%s/%s", host, namespace, moduleName, target)

	versions, err := m.fetchRegistryModuleVersions(moduleURL)
	if err != nil {
		return nil, err
	}

	if len(versions) == 0 {
		return nil, errors.New("No versions found for registry module")
	}

	// We now have a list of versions for the module, so we need to find the latest matching version
	constraints, err := goversion.NewConstraint(moduleCall.Version)
	if err != nil {
		return nil, err
	}

	var matchingVersion string

	for _, rawVersion := range versions {
		version, err := goversion.NewVersion(rawVersion)
		if err != nil {
			return nil, err
		}

		if constraints.Check(version) {
			matchingVersion = rawVersion
			break
		}
	}

	return &RegistryModuleCheckResult{
		Version:     matchingVersion,
		DownloadURL: fmt.Sprintf("%s/%s/download", moduleURL, matchingVersion),
	}, nil
}

func (m *ModuleLoader) fetchRegistryModuleVersions(moduleURL string) ([]string, error) {
	httpClient := &http.Client{}
	resp, err := httpClient.Get(moduleURL + "/versions")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch registry module versions")
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
		return nil, errors.Wrap(err, "Failed to read module versions response")
	}

	err = json.Unmarshal(respBody, &versionsResp)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal module versions response")
	}

	if len(versionsResp.Modules) == 0 {
		return nil, fmt.Errorf("Module versions endpoint returned no modules")
	}

	versions := make([]string, 0, len(versionsResp.Modules[0].Versions))

	for _, v := range versionsResp.Modules[0].Versions {
		versions = append(versions, v.Version)
	}

	// TODO: Do we need to sort the versions or are they always sorted already?

	return versions, nil
}

func (m *ModuleLoader) downloadRegistryModule(downloadURL string, dest string) error {
	httpClient := &http.Client{}
	resp, err := httpClient.Get(downloadURL)
	if err != nil {
		return errors.Wrap(err, "Failed to download registry module")
	}

	source := resp.Header.Get("X-Terraform-Get")
	if source == "" {
		return errors.New("download URL has no X-Terraform-Get header")
	}

	return m.downloadRemoteModule(source, dest)
}

func (m *ModuleLoader) downloadRemoteModule(source string, dest string) error {
	client := getter.Client{
		Src:  source,
		Dst:  path.Join(m.path, dest),
		Pwd:  path.Join(m.path, dest),
		Mode: getter.ClientModeDir,
		// TODO: check these:
		// https://github.com/hashicorp/terraform/blob/affe2c329561f40f13c0e94f4570321977527a77/internal/getmodules/getter.go#L57
	}

	return client.Get()
}
