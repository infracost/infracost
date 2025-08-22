package modules

import (
	"bytes"
	"crypto/md5" // nolint
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	giturls "github.com/chainguard-dev/git-urls"
	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/infracost/infracost/internal/metrics"
	"github.com/otiai10/copy"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	intSync "github.com/infracost/infracost/internal/sync"
)

var (
	// downloadDir is name of the directory where remote modules are download
	downloadDir = ".infracost/terraform_modules"
	// manifestPath is the name of the module manifest file which stores the metadata of the modules
	manifestPath = ".infracost/terraform_modules/manifest.json"
	// tfManifestPath is the name of the terraform module manifest file which stores the metadata of the modules
	tfManifestPath = ".terraform/modules/modules.json"

	supportedManifestVersion = "2.0"

	// maxSparseCheckoutDepth is the maximum depth to which we will follow symlinks when adding directories
	// to the sparse-checkout file list. This is currently set to a low value, since increasing it is likely
	// to cause performance issues. If we need to increase it in the future, we should consider adding an
	// option to allow users to set this value.
	maxSparseCheckoutDepth = 1
)

// RemoteCache is an interface that defines the methods for a remote cache, i.e. an S3 bucket.
type RemoteCache interface {
	Exists(key string, public bool) (bool, error)
	Get(key string, dest string, public bool) error
	Put(key string, src string, ttl time.Duration, public bool) error
}

// PublicModuleChecker is an interface that defines the method for checking if a module is public.
type PublicModuleChecker interface {
	IsPublicModule(moduleAddr string) (bool, error)
}

// ModuleLoader handles the loading of Terraform modules. It supports local, registry and other remote modules.
//
// The path should be the root directory of the Terraform project. We use a distinct module loader per Terraform project,
// because at the moment the cache is per project. The cache reads the manifest.json file from the path's
// .infracost/terraform_modules directory. We could implement a global cache in the future, but for now have decided
// to go with the same approach as Terraform.
type ModuleLoader struct {
	// cachePath is the path to the directory that Infracost will download modules to.
	// This is normally the top level directory of a multi-project environment, where the
	// Infracost config file resides or project auto-detection starts from.
	cachePath      string
	cache          *Cache
	sync           *intSync.KeyMutex
	hclParser      *SharedHCLParser
	sourceMap      config.TerraformSourceMap
	sourceMapRegex config.TerraformSourceMapRegex

	packageFetcher *PackageFetcher
	registryLoader *RegistryLoader
	logger         zerolog.Logger
}

type SourceMapResult struct {
	Source   string
	Version  string
	RawQuery string
}
type ModuleLoaderOptions struct {
	CachePath           string
	HCLParser           *SharedHCLParser
	CredentialsSource   *CredentialsSource
	SourceMap           config.TerraformSourceMap
	SourceMapRegex      config.TerraformSourceMapRegex
	Logger              zerolog.Logger
	ModuleSync          *intSync.KeyMutex
	RemoteCache         RemoteCache
	PublicModuleChecker PublicModuleChecker
}

// NewModuleLoader constructs a new module loader
func NewModuleLoader(opts ModuleLoaderOptions) *ModuleLoader {
	fetcher := NewPackageFetcher(opts.RemoteCache, opts.Logger, WithPublicModuleChecker(opts.PublicModuleChecker))
	// we need to have a disco for each project that has defined credentials
	d := NewDisco(opts.CredentialsSource, opts.Logger)

	if err := opts.SourceMapRegex.Compile(); err != nil {
		opts.Logger.Error().Err(err).Msg("error compiling source map regex")
	}

	m := &ModuleLoader{
		cachePath:      opts.CachePath,
		cache:          NewCache(d, opts.Logger),
		hclParser:      opts.HCLParser,
		sourceMap:      opts.SourceMap,
		sourceMapRegex: opts.SourceMapRegex,
		packageFetcher: fetcher,
		logger:         opts.Logger,
		sync:           opts.ModuleSync,
	}

	m.registryLoader = NewRegistryLoader(fetcher, d, opts.Logger)

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
	sum := md5.Sum([]byte(rel)) // nolint
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

	module, err := m.loadModuleFromPath(path)
	if err != nil {
		return nil, err
	}

	numJobs := len(module.ModuleCalls)
	jobs := make(chan *tfconfig.ModuleCall, numJobs)
	manifestModules := make([]*ManifestModule, len(jobs))
	for _, moduleCall := range module.ModuleCalls {
		jobs <- moduleCall
	}
	close(jobs)

	errGroup := &errgroup.Group{}
	manifestMu := sync.Mutex{}

	remoteModuleCounter := metrics.GetCounter("remote_module.count", true)

	for i := 0; i < getProcessCount(); i++ {
		errGroup.Go(func() error {
			for moduleCall := range jobs {
				metadata, err := m.loadModule(moduleCall, path, prefix)
				if err != nil {
					return err
				}

				// only include non-local modules in the manifest since we don't want to cache local ones.
				if !IsLocalModule(metadata.Source) {
					remoteModuleCounter.Inc()
					manifestMu.Lock()
					manifestModules = append(manifestModules, metadata)
					manifestMu.Unlock()
				}

				if IsLocalModule(metadata.Source) && metadata.IsSourceMapped {
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
	isSourceMapped := false

	mappedResult, err := m.MapSourceWithRegex(source)
	if err != nil {
		return nil, err
	}

	if mappedResult.Source != source {
		m.logger.Debug().Msgf("remapping module source %s to %s", source, mappedResult.Source)
		source = mappedResult.Source
		isSourceMapped = true
	}

	if mappedResult.Version != "" {
		m.logger.Debug().Msgf("remapping module version %s to %s", version, mappedResult.Version)
		version = mappedResult.Version
		isSourceMapped = true
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
	if IsLocalModule(source) {
		dir, err := m.cachePathRel(filepath.Join(parentPath, source))
		if err != nil {
			return nil, err
		}

		m.logger.Debug().Msgf("loading local module %s from %s", key, dir)

		// If the module is in a git repo we need to check that it has been checked out.
		// If it hasn't been checked out then we need to add it to the sparse-checkout file list.
		m.logger.Trace().Msgf("finding git repo root for path %s", parentPath)
		repoRoot, err := findGitRepoRoot(parentPath)
		if err == nil {
			// Get the dir relative to the repoRoot
			absDir, err := filepath.Abs(dir)
			if err != nil {
				return nil, err
			}

			relDir, err := filepath.Rel(repoRoot, absDir)
			if err != nil {
				return nil, err
			}

			err = m.checkoutPathIfRequired(repoRoot, relDir)
			if err != nil {
				return nil, err
			}
		}

		return &ManifestModule{
			Key:            key,
			Source:         source,
			Dir:            path.Clean(dir),
			IsSourceMapped: isSourceMapped,
		}, nil
	}

	moduleLoadTimer := metrics.GetTimer("submodule.remote_load.duration", false, source, version).Start()
	defer moduleLoadTimer.Stop()

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

// checkoutPathIfRequired checks if the given directories are in the sparse-checkout file list and adds them if not.
func (m *ModuleLoader) checkoutPathIfRequired(repoRoot string, dir string) error {
	// Lock the git repo root so we don't have multiple calls trying to read and update the sparse-checkout file list.
	unlock := m.sync.Lock(repoRoot)
	defer unlock()

	// Check if sparse-checkout is enabled for the repo
	enabled, err := isSparseCheckoutEnabled(repoRoot)
	if err != nil {
		return err
	}

	if !enabled {
		m.logger.Trace().Msgf("sparse-checkout not enabled for path %s", repoRoot)
		return nil
	}

	// Get the list of sparse checkout directories (and assume sparse checkout is enabled if this succeeds)
	m.logger.Trace().Msgf("getting sparse checkout directories for path %s", repoRoot)
	existingDirs, err := getSparseCheckoutDirs(repoRoot)
	if err != nil {
		// If the error indicates that sparse checkout is not enabled, just return nil
		// Even though we check this above, we need to check it again here because the sparse-checkout
		// config might be enabled but sparse-checkout might not be fully initialized
		if err.Error() == "sparse-checkout not enabled" {
			m.logger.Trace().Msgf("sparse-checkout not enabled for path %s", repoRoot)
			return nil
		}
	}

	sourceURL, err := getGitURL(repoRoot)
	if err != nil {
		return err
	}

	mu := &sync.Mutex{}

	return RecursivelyAddDirsToSparseCheckout(repoRoot, sourceURL, m.packageFetcher, existingDirs, []string{dir}, mu, m.logger, 0)
}

// RecursivelyAddDirsToSparseCheckout adds the given directories to the sparse-checkout file list.
// It then checks any symlinks within the directories and adds them to the sparse-checkout file list as well.
func RecursivelyAddDirsToSparseCheckout(repoRoot string, sourceURL string, packageFetcher *PackageFetcher, existingDirs []string, dirs []string, mu *sync.Mutex, logger zerolog.Logger, depth int) error {
	newDirs := make([]string, 0, len(dirs))

	// Sort the existing directories and dirs to be added by length
	// This ensures that parent directories are added before child directories
	// since they cover the child directories anyway.
	sort.Slice(existingDirs, func(i, j int) bool {
		return len(existingDirs[i]) < len(existingDirs[j])
	})
	sort.Slice(dirs, func(i, j int) bool {
		return len(dirs[i]) < len(dirs[j])
	})

	for _, dir := range dirs {
		if isCoveredByExistingDirs(existingDirs, dir) {
			continue
		}

		existingDirs = append(existingDirs, dir)
		newDirs = append(newDirs, dir)
	}
	if len(newDirs) == 0 {
		return nil
	}

	parsedSourceURL, err := url.Parse(sourceURL)
	if err != nil {
		return err
	}

	// Create a temporary directory for this fetch
	tmpDir, err := os.MkdirTemp(os.TempDir(), "infracost-sparse-checkout")
	if err != nil {
		return fmt.Errorf("error creating temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	logger.Trace().Msgf("adding dirs to sparse-checkout for repo %s: %v", repoRoot, newDirs)
	for _, dir := range newDirs {
		q := parsedSourceURL.Query()
		q.Set("subdir", dir)
		parsedSourceURL.RawQuery = q.Encode()
		s := parsedSourceURL.String()

		// Load the package either from the cache or by pulling from the remote.
		// We load it into a temporary directory so we can then merge it into the repo root.
		// If we load it into the repo root then the fetcher will try and cache
		// the entire repo root which we don't want.
		dirTmpDir := filepath.Join(tmpDir, "fetch-"+filepath.Base(dir))
		err := packageFetcher.Fetch(s, dirTmpDir)
		if err != nil {
			return fmt.Errorf("error fetching module %s: %w", s, err)
		}

		mu.Lock()
		// Copy the downloaded package into the repo root
		opt := copy.Options{
			OnSymlink: func(src string) copy.SymlinkAction {
				return copy.Shallow
			},
		}
		err = copy.Copy(dirTmpDir, repoRoot, opt)
		if err != nil {
			return fmt.Errorf("error copying module %s to repo root: %w", s, err)
		}

		// After we've fetched the package we need to update the sparse checkout list
		// because the package fetcher retrieved it from the remote cache this won't
		// have been done and future calls to download other subdirs won't work
		err = setSparseCheckoutDirs(repoRoot, existingDirs)
		mu.Unlock()

		if err != nil {
			return fmt.Errorf("error setting sparse checkout list: %w", err)
		}
	}

	if depth >= maxSparseCheckoutDepth {
		return nil
	}

	var additionalDirs []string
	for _, dir := range newDirs {
		symlinkedDirs, err := ResolveSymLinkedDirs(repoRoot, dir)
		if err != nil {
			return fmt.Errorf("error resolving symlinks for dir %s: %w", dir, err)
		}

		for _, symlinkedDir := range symlinkedDirs {
			if !isCoveredByExistingDirs(existingDirs, symlinkedDir) {
				additionalDirs = append(additionalDirs, symlinkedDir)
			}
		}
	}

	if len(additionalDirs) > 0 {
		logger.Trace().Msgf("recursively adding symlinked dirs to sparse-checkout for repo %s: %v", repoRoot, additionalDirs)
		return RecursivelyAddDirsToSparseCheckout(repoRoot, sourceURL, packageFetcher, existingDirs, additionalDirs, mu, logger, depth+1)
	}

	return nil
}

// isCoveredByExistingDirs checks if the given directory is covered by any of the existing directories
// i.e. if it is a subdirectory of any of the existing directories.
func isCoveredByExistingDirs(existingDirs []string, dir string) bool {
	for _, existingDir := range existingDirs {
		if dir == existingDir || strings.HasPrefix(dir, existingDir+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

// findGitRepoRoot finds the root of the Git repository given a starting path
func findGitRepoRoot(startPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = startPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository")
	}
	return strings.TrimSpace(string(output)), nil
}

// getGitURL gets the Git URL for the given path
func getGitURL(path string) (string, error) {
	// Get remote
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error getting git URL: %w", err)
	}

	remote := strings.TrimSpace(string(output))

	// Get commit
	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = path
	output, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error getting git commit: %w", err)
	}

	commit := strings.TrimSpace(string(output))

	return fmt.Sprintf("git::%s?ref=%s", remote, commit), nil
}

// isSparseCheckoutEnabled checks if sparse-checkout is enabled in the repository
func isSparseCheckoutEnabled(repoRoot string) (bool, error) {
	cmd := exec.Command("git", "config", "--get", "core.sparseCheckout")
	cmd.Dir = repoRoot
	output, err := cmd.Output()
	if err != nil {
		// if exit status is 1, then the config is not set
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return false, nil
		}

		return false, fmt.Errorf("error checking if sparse-checkout is enabled: %w", err)
	}
	return string(output) == "true\n", nil
}

// getSparseCheckoutDirs gets the list of directories currently in sparse-checkout
func getSparseCheckoutDirs(repoRoot string) ([]string, error) {
	cmd := exec.Command("git", "sparse-checkout", "list")
	cmd.Dir = repoRoot

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if strings.Contains(stderrStr, "this worktree is not sparse") {
			return nil, errors.New("sparse-checkout not enabled")
		}

		return nil, fmt.Errorf("error getting sparse-checkout list: %w", err)
	}

	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return nil, nil
	}

	return strings.Split(output, "\n"), nil
}

// setSparseCheckoutDirs sets the sparse-checkout list to include the given directories
func setSparseCheckoutDirs(repoRoot string, dirs []string) error {
	cmd := exec.Command("git", "sparse-checkout", "set")
	cmd.Dir = repoRoot
	cmd.Args = append(cmd.Args, dirs...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error setting sparse-checkout list: %w", err)
	}

	return nil
}

// ResolveSymLinkedDirs traverses the directory to find symlinks and resolve their destinations
func ResolveSymLinkedDirs(repoRoot, dir string) ([]string, error) {
	dirsToInclude := make(map[string]struct{})
	basePath := path.Join(repoRoot, dir)

	err := filepath.Walk(basePath, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking the path: %w", err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			resolvedPath, err := resolveSymlink(currentPath)
			if err != nil {
				return fmt.Errorf("error resolving symlink: %w", err)
			}

			// Ensure the resolved path is within the repoRoot
			if !strings.HasPrefix(resolvedPath, repoRoot) {
				return nil
			}

			resolvedDir := filepath.Dir(resolvedPath)

			resolvedRelDir, err := filepath.Rel(repoRoot, resolvedDir)
			if err != nil {
				return fmt.Errorf("error getting relative path: %w", err)
			}

			// Skip if it's already included or within the original directory
			if _, alreadyIncluded := dirsToInclude[resolvedRelDir]; alreadyIncluded || strings.HasPrefix(resolvedRelDir, dir) {
				return nil
			}

			dirsToInclude[resolvedRelDir] = struct{}{}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	dirs := make([]string, 0, len(dirsToInclude))
	for d := range dirsToInclude {
		dirs = append(dirs, d)
	}

	return dirs, nil
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

	// lock the download destination so that we don't interact with an incomplete download.
	// we can't use module address as the key here because the module address might be different for the same module,
	// e.g. ssh vs https
	unlock := m.sync.Lock(dest)
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

	// If sparse checkout is enabled add the subdir to the module URL as a query param
	// so go-getter only downloads the required directory.
	if os.Getenv("INFRACOST_SPARSE_CHECKOUT") == "true" {
		u, err := url.Parse(moduleAddr)
		if err != nil {
			return nil, err
		}
		q := u.Query()
		q.Set("subdir", submodulePath)
		u.RawQuery = q.Encode()
		moduleAddr = u.String()
	}

	dest := m.downloadDest(moduleAddr, "")
	moduleDownloadDir, err := m.cachePathRel(dest)
	if err != nil {
		return nil, err
	}
	manifestModule.Dir = path.Clean(filepath.Join(moduleDownloadDir, submodulePath))

	// lock the download destination so that we don't interact with an incomplete download.
	// we can't use module address here as the version may be different
	unlock := m.sync.Lock(dest)
	defer unlock()

	_, err = os.Stat(dest)
	if err == nil {
		return manifestModule, nil
	}

	err = m.packageFetcher.Fetch(moduleAddr, dest)
	if err != nil {
		_ = os.RemoveAll(dest)
		return nil, schema.NewFailedDownloadDiagnostic(source, err)
	}

	return manifestModule, nil
}

func (m *ModuleLoader) downloadDest(moduleAddr string, version string) string {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(moduleAddr+version))) // nolint
	return filepath.Join(m.downloadDir(), hash)
}

func (m *ModuleLoader) cachePathRel(targetPath string) (string, error) {
	rel, relerr := filepath.Rel(m.cachePath, targetPath)
	if relerr == nil {
		return rel, nil
	}
	m.logger.Debug().Msgf("Failed to filepath.Rel cache=%s target=%s: %v", m.cachePath, targetPath, relerr)

	// try converting to absolute paths
	absCachePath, abserr := filepath.Abs(m.cachePath)
	if abserr != nil {
		m.logger.Debug().Msgf("Failed to filepath.Abs cachePath: %v", abserr)
		return "", relerr
	}

	absTargetPath, abserr := filepath.Abs(targetPath)
	if abserr != nil {
		m.logger.Debug().Msgf("Failed to filepath.Abs target: %v", abserr)
		return "", relerr
	}

	m.logger.Debug().Msgf("Attempting filepath.Rel on abs paths cache=%s, target=%s", absCachePath, absTargetPath)
	return filepath.Rel(absCachePath, absTargetPath)
}

// IsLocalModule checks if the module is a local module by checking
// if the module source starts with any known local prefixes
func IsLocalModule(source string) bool {
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

// MapSource maps the module source to a new source if it is in the source map
// otherwise it returns the original source. It works similarly to the
// TERRAGRUNT_SOURCE_MAP environment variable except it matches by prefixes
// and supports query params. It works by matching the longest prefix first,
// so the most specific prefix is matched first.
//
// It does not support mapping registry versions to git tags since we can't
// guarantee that the tag is correct - depending on the git repo the version
// might be prefixed with a 'v' or not.
func MapSource(sourceMap config.TerraformSourceMap, source string) (SourceMapResult, error) {
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
		// Try parsing it as a git URL
		parsedSourceURL, err = giturls.Parse(moduleAddr)
		if err != nil {
			return SourceMapResult{}, err
		}
	}

	parsedDestURL, err := url.Parse(destSource)
	if err != nil {
		// Try parsing it as a git URL
		parsedDestURL, err = giturls.Parse(destSource)
		if err != nil {
			return SourceMapResult{}, err
		}
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

// MapSourceWithRegex maps the module source using regex patterns if available.
// Falls back to the standard MapSource function if regex mapping is not configured
// or if no regex patterns match.
func (m *ModuleLoader) MapSourceWithRegex(source string) (SourceMapResult, error) {
	if len(m.sourceMapRegex) > 0 {
		mappedSource, err := config.ApplyRegexMapping(m.sourceMapRegex, source)
		if err != nil {
			return SourceMapResult{}, err
		}

		if mappedSource != source {
			// Split for submodule path and extract version if available
			moduleAddr, submodulePath, err := splitModuleSubDir(mappedSource)
			if err != nil {
				return SourceMapResult{}, err
			}

			result := SourceMapResult{
				Source:   mappedSource,
				Version:  "",
				RawQuery: "",
			}

			// Parse the URL to extract ref for version
			parsedURL, err := url.Parse(moduleAddr)
			if err == nil && parsedURL.RawQuery != "" {
				query := parsedURL.Query()
				result.RawQuery = parsedURL.RawQuery

				// If query params have a ref then use that as the version
				ref := query.Get("ref")
				if ref != "" {
					result.Version = strings.TrimPrefix(ref, "v")
				}

				// Reconstruct the URL without query for final source
				parsedURL.RawQuery = ""
				result.Source = joinModuleSubDir(parsedURL.String(), submodulePath)
			}

			return result, nil
		}
	}

	return MapSource(m.sourceMap, source)
}

// resolveSymlink resolves symlinks even if the target does not exist
func resolveSymlink(path string) (string, error) {
	link, err := os.Readlink(path)
	if err != nil {
		return "", err
	}

	if filepath.IsAbs(link) {
		return link, nil
	}

	return filepath.Join(filepath.Dir(path), link), nil
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
