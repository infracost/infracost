package modules

import (
	"crypto/md5" //nolint
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	getter "github.com/hashicorp/go-getter"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	intSync "github.com/infracost/infracost/internal/sync"
	"github.com/infracost/infracost/internal/ui"
)

var (
	// downloadDir is name of the directory where remote modules are download
	downloadDir = ".infracost/terraform_modules"
	// manifestPath is the name of the module manifest file which stores the metadata of the modules
	manifestPath = ".infracost/terraform_modules/manifest.json"
	// tfManifestPath is the name of the terraform module manifest file which stores the metadata of the modules
	tfManifestPath = ".terraform/modules/modules.json"

	supportedManifestVersion = "2.0"
)

// ModuleLoader handles the loading of Terraform modules. It supports local, registry and other remote modules.
//
// The path should be the root directory of the Terraform project. We use a distinct module loader per Terraform project,
// because at the moment the cache is per project. The cache reads the manifest.json file from the path's
// .infracost/terraform_modules directory. We could implement a global cache in the future, but for now have decided
// to go with the same approach as Terraform.
type ModuleLoader struct {
	NewSpinner ui.SpinnerFunc

	// cachePath is the path to the directory that Infracost will download modules to.
	// This is normally the top level directory of a multi-project environment, where the
	// Infracost config file resides or project auto-detection starts from.
	cachePath string
	cache     *Cache
	sync      *intSync.KeyMutex
	hclParser *SharedHCLParser
	sourceMap config.TerraformSourceMap

	packageFetcher *PackageFetcher
	registryLoader *RegistryLoader
	logger         zerolog.Logger
}

type SourceMapResult struct {
	Source   string
	Version  string
	RawQuery string
}

// NewModuleLoader constructs a new module loader
func NewModuleLoader(cachePath string, hclParser *SharedHCLParser, credentialsSource *CredentialsSource, sourceMap config.TerraformSourceMap, logger zerolog.Logger, moduleSync *intSync.KeyMutex) *ModuleLoader {
	fetcher := NewPackageFetcher(logger)
	// we need to have a disco for each project that has defined credentials
	d := NewDisco(credentialsSource, logger)

	m := &ModuleLoader{
		cachePath:      cachePath,
		cache:          NewCache(d, logger),
		hclParser:      hclParser,
		sourceMap:      sourceMap,
		packageFetcher: fetcher,
		logger:         logger,
		sync:           moduleSync,
	}

	m.registryLoader = NewRegistryLoader(fetcher, d, logger)

	return m
}

// downloadDir returns the path to the directory where remote modules are downloaded relative to the current working directory
func (m *ModuleLoader) downloadDir() string {
	return filepath.Join(m.cachePath, downloadDir)
}

// manifestFilePath is the path to the module manifest file relative to the current working directory
func (m *ModuleLoader) manifestFilePath(projectPath string) string {
	if m.cachePath == projectPath {
		return filepath.Join(m.cachePath, manifestPath)
	}

	rel, _ := filepath.Rel(m.cachePath, projectPath)
	sum := md5.Sum([]byte(rel)) //nolint
	return filepath.Join(m.cachePath, ".infracost/terraform_modules/", fmt.Sprintf("manifest-%x.json", sum))
}

// tfManifestFilePath is the path to the Terraform module manifest file relative to the current working directory.
func (m *ModuleLoader) tfManifestFilePath(path string) string {
	return filepath.Join(path, tfManifestPath)
}

// Load loads the modules from the given path.
// For each module it checks if the module has already been downloaded, by checking if iut exists in the manifest
// If not then it downloads the module from the registry or from a remote source and updates the module manifest with the latest metadata.
func (m *ModuleLoader) Load(path string) (man *Manifest, err error) {
	defer func() {
		if man != nil {
			man.cachePath = m.cachePath
		}
	}()

	if m.NewSpinner != nil {
		spin := m.NewSpinner("Downloading Terraform modules")
		defer spin.Success()
	}

	manifest := &Manifest{}
	manifestFilePath := m.manifestFilePath(path)
	_, err = os.Stat(manifestFilePath)
	if errors.Is(err, os.ErrNotExist) {
		m.logger.Debug().Msg("No existing module manifest file found")

		tfManifestFilePath := m.tfManifestFilePath(path)
		_, err = os.Stat(tfManifestFilePath)
		if err == nil {
			manifest, err = readManifest(tfManifestFilePath)
			if err == nil {
				// let's make the module dirs relative to the path directory as later
				// we'll look up the modules based on the cache path at the Infracost root (where the infracost.yml
				// resides or where the --path autodetect started for multi-project)
				for i, module := range manifest.Modules {
					dir := path
					if m.cachePath != "" {
						dir, _ = filepath.Rel(m.cachePath, path)
					}

					manifest.Modules[i].Dir = filepath.Join(dir, module.Dir)
				}

				return manifest, nil
			}

			m.logger.Debug().Err(err).Msg("error reading terraform module manifest")
		}
	} else if err != nil {
		m.logger.Debug().Err(err).Msg("error checking for existing module manifest")
	} else {
		manifest, err = readManifest(manifestFilePath)
		if err != nil {
			m.logger.Debug().Err(err).Msg("could not read module manifest")
		}

		if manifest.Version != supportedManifestVersion {
			manifest = &Manifest{cachePath: m.cachePath}
		}
	}
	m.cache.loadFromManifest(manifest)

	metadatas, err := m.loadModules(path, "")
	if err != nil {
		return nil, err
	}

	manifest.Modules = metadatas
	manifest.Path = path
	manifest.Version = supportedManifestVersion

	err = writeManifest(manifest, manifestFilePath)
	if err != nil {
		m.logger.Debug().Err(err).Msg("error writing module manifest")
	}

	return manifest, nil
}

// loadModules recursively loads the modules from the given path.
func (m *ModuleLoader) loadModules(path string, prefix string) ([]*ManifestModule, error) {
	manifestModules := make([]*ManifestModule, 0)

	module, err := m.loadModuleFromPath(path)
	if err != nil {
		return nil, err
	}

	numJobs := len(module.ModuleCalls)
	jobs := make(chan *tfconfig.ModuleCall, numJobs)
	for _, moduleCall := range module.ModuleCalls {
		jobs <- moduleCall
	}
	close(jobs)

	errGroup := &errgroup.Group{}
	manifestMu := sync.Mutex{}

	for i := 0; i < getProcessCount(); i++ {
		errGroup.Go(func() error {
			for moduleCall := range jobs {
				metadata, err := m.loadModule(moduleCall, path, prefix)
				if err != nil {
					return err
				}

				// only include non-local modules in the manifest since we don't want to cache local ones.
				if !isLocalModule(metadata.Source) {
					manifestMu.Lock()
					manifestModules = append(manifestModules, metadata)
					manifestMu.Unlock()
				}

				moduleDir := filepath.Join(m.cachePath, metadata.Dir)
				nestedManifestModules, err := m.loadModules(moduleDir, metadata.Key+".")
				if err != nil {
					return err
				}

				manifestMu.Lock()
				manifestModules = append(manifestModules, nestedManifestModules...)
				manifestMu.Unlock()
			}

			return nil
		})
	}

	err = errGroup.Wait()
	if err != nil {
		return manifestModules, fmt.Errorf("could not load modules for path %s %w", path, err)
	}

	return manifestModules, nil
}

// loadModule loads the module metadata from the given module call.
// It works by doing the following:
// 1. Checks if the module is already downloaded and the version/source has not changed.
// 2. Checks if the module is a local module.
// 3. Checks if the module is a registry module and downloads it.
// 4. Checks if the module is a remote module and downloads it.
func (m *ModuleLoader) loadModule(moduleCall *tfconfig.ModuleCall, parentPath string, prefix string) (*ManifestModule, error) {
	key := prefix + moduleCall.Name
	source := moduleCall.Source
	version := moduleCall.Version

	mappedResult, err := mapSource(m.sourceMap, source)
	if err != nil {
		return nil, err
	}

	if mappedResult.Source != source {
		m.logger.Debug().Msgf("remapping module source %s to %s", source, mappedResult.Source)
		source = mappedResult.Source
	}

	if mappedResult.Version != "" {
		m.logger.Debug().Msgf("remapping module version %s to %s", version, mappedResult.Version)
		version = mappedResult.Version
	}

	manifestModule, err := m.cache.lookupModule(key, moduleCall)
	if err == nil {
		m.logger.Debug().Msgf("module %s already loaded", key)

		// Test if we can actually load the module. If not, then we should try re-loading it.
		// This can happen if the directory the module was downloaded to has been deleted and moved
		// so the existing manifest.json is out-of-date.
		_, loadModErr := m.loadModuleFromPath(path.Join(m.cachePath, manifestModule.Dir))
		if loadModErr == nil {
			return manifestModule, nil
		}

		m.logger.Debug().Msgf("module %s cannot be loaded, re-loading: %s", key, loadModErr)
	} else {
		m.logger.Debug().Msgf("module %s needs loading: %s", key, err.Error())
	}
	if isLocalModule(source) {
		dir, err := m.cachePathRel(filepath.Join(parentPath, source))
		if err != nil {
			return nil, err
		}

		m.logger.Debug().Msgf("loading local module %s from %s", key, dir)
		return &ManifestModule{
			Key:    key,
			Source: source,
			Dir:    path.Clean(dir),
		}, nil
	}

	manifestModule, err = m.loadRegistryModule(key, source, version)
	if err != nil {
		return nil, err
	}

	if manifestModule != nil {
		return manifestModule, nil
	}

	// For remote modules that have had their source mapped we need to include
	// the query params in the source URL.
	if mappedResult.RawQuery != "" {
		parsedSourceURL, err := url.Parse(source)
		if err != nil {
			return nil, err
		}

		parsedSourceURL.RawQuery = mappedResult.RawQuery
		source = parsedSourceURL.String()
	}

	manifestModule, err = m.loadRemoteModule(key, source)
	if err != nil {
		return nil, err
	}

	return manifestModule, nil
}

func (m *ModuleLoader) loadModuleFromPath(fullPath string) (*tfconfig.Module, error) {
	mod := tfconfig.NewModule(fullPath)

	fileInfos, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	for _, info := range fileInfos {
		if info.IsDir() {
			continue
		}

		var parseFunc func(filename string) (*hcl.File, hcl.Diagnostics)
		if strings.HasSuffix(info.Name(), ".tf") {
			parseFunc = m.hclParser.ParseHCLFile
		}

		if strings.HasSuffix(info.Name(), ".tf.json") {
			parseFunc = m.hclParser.ParseJSONFile
		}

		// this is not a file we can parse:
		if parseFunc == nil {
			continue
		}

		path := filepath.Join(fullPath, info.Name())
		f, fileDiag := parseFunc(path)
		if fileDiag != nil && fileDiag.HasErrors() {
			return nil, fmt.Errorf("failed to parse file %s diag: %w", path, fileDiag)
		}

		if f == nil {
			continue
		}

		contentDiag := tfconfig.LoadModuleFromFile(f, mod)
		if contentDiag != nil && contentDiag.HasErrors() {
			return nil, fmt.Errorf("failed to load module from file %s diag: %w", path, contentDiag)
		}
	}

	return mod, nil
}

func (m *ModuleLoader) loadRegistryModule(key string, source string, version string) (*ManifestModule, error) {
	manifestModule := &ManifestModule{
		Key: key,
	}

	moduleAddr, submodulePath, err := splitModuleSubDir(source)
	if err != nil {
		return nil, err
	}

	dest := m.downloadDest(moduleAddr, version)
	moduleDownloadDir, err := m.cachePathRel(dest)
	if err != nil {
		return nil, err
	}
	manifestModule.Dir = path.Clean(filepath.Join(moduleDownloadDir, submodulePath))

	// lock the module address so that we don't interact with an incomplete download.
	unlock := m.sync.Lock(moduleAddr)
	defer unlock()

	lookupResult, err := m.registryLoader.lookupModule(moduleAddr, version)
	if err != nil {
		return nil, schema.NewPrivateRegistryDiag(source, nil, err)
	}

	if lookupResult.OK {
		// The source might not have the registry hostname if it is using the default registry
		// so we set the source here to the lookup result's source which always includes the registry hostname.
		manifestModule.Source = joinModuleSubDir(lookupResult.ModuleURL.RawSource, submodulePath)
		manifestModule.DownloadURL, _ = m.registryLoader.DownloadLocation(lookupResult.ModuleURL, lookupResult.Version)
		manifestModule.Version = lookupResult.Version

		_, err = os.Stat(dest)
		if err == nil {
			return manifestModule, nil
		}

		err = m.registryLoader.downloadModule(lookupResult, dest)
		if err != nil {
			return nil, schema.NewPrivateRegistryDiag(source, strPtr(lookupResult.ModuleURL.Location), err)
		}

		return manifestModule, nil
	}

	return nil, nil
}

func strPtr(s string) *string {
	return &s
}

func (m *ModuleLoader) loadRemoteModule(key string, source string) (*ManifestModule, error) {
	manifestModule := &ManifestModule{
		Key:    key,
		Source: source,
	}

	moduleAddr, submodulePath, err := splitModuleSubDir(source)
	if err != nil {
		return nil, err
	}

	dest := m.downloadDest(moduleAddr, "")
	moduleDownloadDir, err := m.cachePathRel(dest)
	if err != nil {
		return nil, err
	}
	manifestModule.Dir = path.Clean(filepath.Join(moduleDownloadDir, submodulePath))

	// lock the module address so that we don't interact with an incomplete download.
	unlock := m.sync.Lock(moduleAddr)
	defer unlock()

	_, err = os.Stat(dest)
	if err == nil {
		return manifestModule, nil
	}

	err = m.packageFetcher.fetch(moduleAddr, dest)
	if err != nil {
		return nil, schema.NewFailedDownloadDiagnostic(source, err)
	}

	return manifestModule, nil
}

func (m *ModuleLoader) downloadDest(moduleAddr string, version string) string {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(moduleAddr+version))) //nolint
	return filepath.Join(m.downloadDir(), hash)
}

func (m *ModuleLoader) cachePathRel(targetPath string) (string, error) {
	rel, relerr := filepath.Rel(m.cachePath, targetPath)
	if relerr == nil {
		return rel, nil
	}
	m.logger.Info().Msgf("Failed to filepath.Rel cache=%s target=%s: %v", m.cachePath, targetPath, relerr)

	// try converting to absolute paths
	absCachePath, abserr := filepath.Abs(m.cachePath)
	if abserr != nil {
		m.logger.Info().Msgf("Failed to filepath.Abs cachePath: %v", abserr)
		return "", relerr
	}

	absTargetPath, abserr := filepath.Abs(targetPath)
	if abserr != nil {
		m.logger.Info().Msgf("Failed to filepath.Abs target: %v", abserr)
		return "", relerr
	}

	m.logger.Info().Msgf("Attempting filepath.Rel on abs paths cache=%s, target=%s", absCachePath, absTargetPath)
	return filepath.Rel(absCachePath, absTargetPath)
}

// isLocalModule checks if the module is a local module by checking
// if the module source starts with any known local prefixes
func isLocalModule(source string) bool {
	return strings.HasPrefix(source, "./") ||
		strings.HasPrefix(source, "../") ||
		strings.HasPrefix(source, ".\\") ||
		strings.HasPrefix(source, "..\\")
}

func splitModuleSubDir(moduleSource string) (string, string, error) {
	moduleAddr, submodulePath := getter.SourceDirSubdir(moduleSource)
	if strings.HasPrefix(submodulePath, "../") {
		return "", "", fmt.Errorf("invalid submodule path '%s'", submodulePath)
	}

	return moduleAddr, submodulePath, nil
}

func joinModuleSubDir(moduleAddr string, submodulePath string) string {
	if submodulePath != "" {
		parsedURL, err := url.Parse(moduleAddr)
		if err != nil || parsedURL.RawQuery == "" {
			return fmt.Sprintf("%s//%s", moduleAddr, submodulePath)
		}

		query := parsedURL.RawQuery
		if query != "" {
			query = "?" + query
		}

		parsedURL.RawQuery = ""
		return fmt.Sprintf("%s//%s%s", parsedURL.String(), submodulePath, query)
	}

	return moduleAddr
}

// mapSource maps the module source to a new source if it is in the source map
// otherwise it returns the original source. It works similarly to the
// TERRAGRUNT_SOURCE_MAP environment variable except it matches by prefixes
// and supports query params. It works by matching the longest prefix first,
// so the most specific prefix is matched first.
//
// It does not support mapping registry versions to git tags since we can't
// guarantee that the tag is correct - depending on the git repo the version
// might be prefixed with a 'v' or not.
func mapSource(sourceMap config.TerraformSourceMap, source string) (SourceMapResult, error) {
	result := SourceMapResult{
		Source:   source,
		Version:  "",
		RawQuery: "",
	}

	if sourceMap == nil {
		return result, nil
	}

	moduleAddr, submodulePath, err := splitModuleSubDir(source)
	if err != nil {
		return SourceMapResult{}, err
	}

	destSource := ""

	// sort the sourceMap keys by length so that we can match the longest prefix first
	// this is important because we want to match the most specific prefix first
	sourceMapKeys := make([]string, 0, len(sourceMap))
	for k := range sourceMap {
		sourceMapKeys = append(sourceMapKeys, k)
	}

	sort.Slice(sourceMapKeys, func(i, j int) bool {
		return len(sourceMapKeys[i]) > len(sourceMapKeys[j])
	})

	for _, prefix := range sourceMapKeys {
		if strings.HasPrefix(moduleAddr, prefix) {
			withoutPrefix := strings.TrimPrefix(moduleAddr, prefix)
			mapped := sourceMap[prefix] + withoutPrefix
			destSource = joinModuleSubDir(mapped, submodulePath)
			break
		}
	}

	// If no result is found
	if destSource == "" {
		return result, nil
	}

	// Merge the query params from the source and dest URLs
	parsedSourceURL, err := url.Parse(moduleAddr)
	if err != nil {
		return SourceMapResult{}, err
	}

	parsedDestURL, err := url.Parse(destSource)
	if err != nil {
		return SourceMapResult{}, err
	}

	sourceQuery := parsedSourceURL.Query()
	destQuery := parsedDestURL.Query()

	for k, v := range sourceQuery {
		if _, ok := destQuery[k]; !ok {
			destQuery[k] = v
		}
	}

	parsedDestURL.RawQuery = ""
	result.Source = parsedDestURL.String()
	result.RawQuery = destQuery.Encode()

	// If the query params have a ref then we should use that as the version
	// for registry modules.
	ref := destQuery.Get("ref")
	if ref != "" {
		result.Version = strings.TrimPrefix(ref, "v")
	}

	return result, nil
}

func getProcessCount() int {
	numWorkers := 4
	numCPU := runtime.NumCPU()

	if numCPU*4 > numWorkers {
		numWorkers = numCPU * 4
	}
	if numWorkers > 16 {
		numWorkers = 16
	}

	return numWorkers
}
