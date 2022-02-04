package hcl

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	getter "github.com/hashicorp/go-getter"
	goversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var modulesCacheDir = ".infracost/terraform_modules"
var moduleManifestFile = ".infracost/terraform_modules/manifest.json"

var validRegistryName = regexp.MustCompile("^[0-9A-Za-z-_]+$")

type ModuleMetadata struct {
	Key     string `json:"Key"`
	Source  string `json:"Source"`
	Version string `json:"Version"`
	Dir     string `json:"Dir"`
}

// ModulesMetadata is a struct that represents the JSON found in the manifest.json file in the .infracost dir
type ModulesManifest struct {
	Modules []ModuleMetadata `json:"Modules"`
}

type RegistryModuleCheckResult struct {
	Version     string
	DownloadURL string
}

type ModuleLoader struct {
	path string
}

func (m *ModuleLoader) cacheDir() string {
	return filepath.Join(m.path, modulesCacheDir)
}

func (m *ModuleLoader) manifestFilePath() string {
	return filepath.Join(m.path, moduleManifestFile)
}

func (m *ModuleLoader) Load() (*ModulesManifest, error) {
	var manifest *ModulesManifest

	_, err := os.Stat(m.manifestFilePath())
	if errors.Is(err, os.ErrNotExist) {
		log.Debugf("No existing module manifest file found")
	} else if err != nil {
		log.Debugf("Error checking for existing module manifest: %s", err)
	} else {
		manifest, err = m.readManifest()
		if err != nil {
			log.Debugf("Error reading module manifest: %s", err)
		}
	}

	if manifest == nil {
		manifest = &ModulesManifest{}
	}

	// Create a cache of modules so we don't have to download them again
	moduleCache := map[string]ModuleMetadata{}
	for _, module := range manifest.Modules {
		moduleCache[module.Key] = module
	}

	metadatas, err := m.loadModules(m.path, moduleCache)
	if err != nil {
		return manifest, nil
	}

	manifest.Modules = metadatas

	err = m.writeManifest(manifest)
	if err != nil {
		log.Debugf("Error writing module manifest: %s", err)
	}

	return manifest, nil
}

func (m *ModuleLoader) loadModules(path string, moduleCache map[string]ModuleMetadata) ([]ModuleMetadata, error) {
	metadatas := make([]ModuleMetadata, 0)

	module, diags := tfconfig.LoadModule(path)
	if diags.HasErrors() {
		return metadatas, diags.Err()
	}

	for _, moduleCall := range module.ModuleCalls {
		metadata, err := m.loadModuleFromCache(moduleCall, moduleCache)
		if err == nil {
			log.Debugf("Module %s already loaded", moduleCall.Name)
		} else {
			log.Debugf("Module %s needs loaded: %s", moduleCall.Name, err.Error())

			metadata, err = m.loadModule(moduleCall, path)
			if err != nil {
				return metadatas, err
			}

			moduleCache[moduleCall.Name] = metadata
		}

		metadatas = append(metadatas, metadata)

		nestedMetadatas, err := m.loadModules(filepath.Join(m.path, metadata.Dir), moduleCache)
		if err != nil {
			return metadatas, err
		}

		metadatas = append(metadatas, nestedMetadatas...)
	}

	return metadatas, nil
}

func (m *ModuleLoader) loadModuleFromCache(moduleCall *tfconfig.ModuleCall, moduleCache map[string]ModuleMetadata) (ModuleMetadata, error) {
	metadata, ok := moduleCache[moduleCall.Name]

	if !ok {
		return metadata, errors.New("not in cache")
	}

	if metadata.Source != moduleCall.Source {
		return metadata, errors.New("source has changed")
	}

	if moduleCall.Version != "" && metadata.Version != "" {
		constraints, err := goversion.NewConstraint(moduleCall.Version)
		if err != nil {
			return metadata, errors.Wrap(err, "invalid version constraint")
		}

		version, err := goversion.NewVersion(metadata.Version)
		if err != nil {
			return metadata, errors.Wrap(err, "invalid version")
		}

		if !constraints.Check(version) {
			return metadata, errors.New("version constraint doesn't match")
		}
	}

	return metadata, nil
}

func (m *ModuleLoader) loadModule(moduleCall *tfconfig.ModuleCall, parentPath string) (ModuleMetadata, error) {
	metadata := ModuleMetadata{
		Key:    moduleCall.Name,
		Source: moduleCall.Source,
	}

	if m.isLocalModule(moduleCall) {
		var err error
		metadata.Dir, err = filepath.Rel(m.path, filepath.Join(parentPath, moduleCall.Source))
		return metadata, err
	}

	if checkResult, err := m.checkRegistryModule(moduleCall); err == nil {
		// TODO: figure out a better path for this to support:
		// - nested modules with the same name
		// - modules with the same source not needing downloaded twice
		dir := filepath.Join(m.cacheDir(), moduleCall.Name)

		err := m.downloadRegistryModule(checkResult.DownloadURL, dir)
		if err != nil {
			return metadata, err
		}

		metadata.Version = checkResult.Version
		metadata.Dir, err = filepath.Rel(m.path, dir)
		return metadata, err
	}

	// TODO: figure out a better path for this to support:
	// - nested modules with the same name
	// - modules with the same source not needing downloaded twice
	dir := filepath.Join(m.cacheDir(), moduleCall.Name)

	err := m.downloadRemoteModule(moduleCall.Source, dir)
	if err != nil {
		return metadata, err
	}

	metadata.Dir, err = filepath.Rel(m.path, dir)
	return metadata, err
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
	defer resp.Body.Close()

	source := resp.Header.Get("X-Terraform-Get")
	if source == "" {
		return errors.New("download URL has no X-Terraform-Get header")
	}

	return m.downloadRemoteModule(source, dest)
}

func (m *ModuleLoader) downloadRemoteModule(source string, dest string) error {
	client := getter.Client{
		Src:  source,
		Dst:  dest,
		Pwd:  dest,
		Mode: getter.ClientModeDir,
		// TODO: check these:
		// https://github.com/hashicorp/terraform/blob/affe2c329561f40f13c0e94f4570321977527a77/internal/getmodules/getter.go#L57
	}

	return client.Get()
}

func (m *ModuleLoader) readManifest() (*ModulesManifest, error) {
	var manifest ModulesManifest

	data, err := os.ReadFile(m.manifestFilePath())
	if err != nil {
		return &manifest, errors.Wrap(err, "Failed to read module manifest")
	}

	err = json.Unmarshal(data, &manifest)
	if err != nil {
		return &manifest, errors.Wrap(err, "Failed to unmarshal module manifest")
	}

	return &manifest, err
}

func (m *ModuleLoader) writeManifest(manifest *ModulesManifest) error {
	b, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return errors.Wrap(err, "Failed to marshal manifest")
	}

	err = os.MkdirAll(filepath.Dir(m.manifestFilePath()), os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "Failed to create directories for manifest")
	}

	err = ioutil.WriteFile(m.manifestFilePath(), b, 0644) // nolint:gosec
	if err != nil {
		return errors.Wrap(err, "Failed to write manifest")
	}

	return nil
}
