package modules

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	tgterraform "github.com/gruntwork-io/terragrunt/terraform"
	getter "github.com/hashicorp/go-getter"
	"github.com/otiai10/copy"
	"github.com/rs/zerolog"

	"github.com/infracost/infracost/internal/logging"
)

var tagRegex = regexp.MustCompile(`^v?\d+\.\d+\.\d+`)
var commitRegex = regexp.MustCompile(`^([0-9a-f]{40})|([0-9a-f]{7})$`)
var defaultTTL = 24 * time.Hour
var tagCommitTTL = 30 * 24 * time.Hour

// PackageFetcher downloads modules from a remote source to the given destination
// This supports all the non-local and non-Terraform registry sources listed here: https://www.terraform.io/language/modules/sources
type PackageFetcher struct {
	localCache  sync.Map
	remoteCache RemoteCache
	logger      zerolog.Logger
}

// NewPackageFetcher constructs a new package fetcher
func NewPackageFetcher(remoteCache RemoteCache, logger zerolog.Logger) *PackageFetcher {
	return &PackageFetcher{
		remoteCache: remoteCache,
		logger:      logger,
	}
}

// fetch downloads the remote module using the go-getter library
// See: https://github.com/hashicorp/go-getter
func (p *PackageFetcher) Fetch(moduleAddr string, dest string) error {
	fetched, err := p.fetchFromLocalCache(moduleAddr, dest)
	if fetched {
		p.logger.Trace().Msgf("cache hit (local): %s", moduleAddr)
		p.logger.Info().Msgf("cache hit (local): %s", moduleAddr)
		return nil
	}

	if err != nil {
		p.logger.Warn().Msgf("error fetching module %s from local cache: %s", moduleAddr, err)
	}

	fetched, err = p.fetchFromRemoteCache(moduleAddr, dest)
	if fetched {
		p.logger.Trace().Msgf("cache hit (remote): %s", moduleAddr)
		p.localCache.Store(moduleAddr, dest)
		return nil
	}

	if err != nil {
		p.logger.Warn().Msgf("error fetching module %s from remote cache: %s", moduleAddr, err)
	}

	p.logger.Trace().Msgf("cache miss: %s", moduleAddr)
	p.logger.Info().Msgf("cache miss: %s", moduleAddr)

	_, err = p.fetchFromRemote(moduleAddr, dest)
	if err != nil {
		return fmt.Errorf("error fetching module %s from remote: %w", moduleAddr, err)
	}

	p.localCache.Store(moduleAddr, dest)

	if p.remoteCache != nil {
		ttl := determineTTL(moduleAddr)
		p.logger.Debug().Msgf("putting module %s into remote cache with ttl %s", moduleAddr, ttl)
		err = p.remoteCache.Put(moduleAddr, dest, ttl)
		if err != nil {
			p.logger.Warn().Msgf("error putting module %s into remote cache: %s", moduleAddr, err)
		}
	}

	return nil
}

func (p *PackageFetcher) fetchFromLocalCache(moduleAddr, dest string) (bool, error) {
	v, ok := p.localCache.Load(moduleAddr)
	if !ok {
		return false, nil
	}

	prevDest, _ := v.(string)

	if prevDest == dest {
		return true, nil
	}

	p.logger.Debug().Msgf("module %s already downloaded, copying from '%s' to '%s'", moduleAddr, prevDest, dest)

	err := os.Mkdir(dest, os.ModePerm)
	if err != nil {
		return false, fmt.Errorf("failed to create directory '%s': %w", dest, err)
	}

	// Skip dotfiles and create new symlinks to be consistent with what Terraform init does
	opt := copy.Options{
		Skip: func(src string) (bool, error) {
			return strings.HasPrefix(filepath.Base(src), "."), nil
		},
		OnSymlink: func(src string) copy.SymlinkAction {
			return copy.Shallow
		},
	}

	err = copy.Copy(prevDest, dest, opt)
	if err != nil {
		return false, fmt.Errorf("failed to copy module from '%s' to '%s': %w", prevDest, dest, err)
	}

	return true, nil
}

func (p *PackageFetcher) fetchFromRemoteCache(moduleAddr, dest string) (bool, error) {
	if p.remoteCache == nil {
		return false, nil
	}

	ok, err := p.remoteCache.Exists(moduleAddr)
	if err != nil {
		return false, err
	}

	if !ok {
		return false, nil
	}

	err = p.remoteCache.Get(moduleAddr, dest)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (p *PackageFetcher) fetchFromRemote(moduleAddr, dest string) (bool, error) {
	decompressors := map[string]getter.Decompressor{}
	for k, decompressor := range getter.Decompressors {
		decompressors[k] = decompressor
	}
	// This one is added by Terraform here: https://github.com/hashicorp/terraform/blob/affe2c329561f40f13c0e94f4570321977527a77/internal/getmodules/getter.go#L64
	// But is not in the list of default compressors here: https://github.com/hashicorp/go-getter/blob/main/decompress.go#L32
	// I'm not sure if we really need it, but added it just in case/
	decompressors["tar.tbz2"] = new(getter.TarBzip2Decompressor)

	getters := make(map[string]getter.Getter, len(getter.Getters))
	for k, g := range getter.Getters {
		getters[k] = g
	}

	// This is a custom getter used by Terragrunt
	getters["tfr"] = &tgterraform.RegistryGetter{}

	getters["git"] = &CustomGitGetter{new(getter.GitGetter)}

	client := getter.Client{
		Src:           moduleAddr,
		Dst:           dest,
		Pwd:           dest,
		Mode:          getter.ClientModeDir,
		Decompressors: decompressors,
		Getters:       getters,
	}

	err := client.Get()
	if err != nil {
		return false, err
	}

	return true, nil
}

func determineTTL(moduleAddr string) time.Duration {
	u, err := url.Parse(moduleAddr)
	if err != nil {
		return defaultTTL
	}

	// Get the ref parameter
	ref := u.Query().Get("ref")
	if ref == "" {
		return defaultTTL
	}

	// Check if ref looks like a git tag or a commit
	isTag := tagRegex.MatchString(ref)
	isCommit := commitRegex.MatchString(ref)

	if isTag || isCommit {
		return tagCommitTTL
	}

	return defaultTTL
}

// CustomGitGetter extends the standard GitGetter transforming SSH sources to
// HTTPs first before attempting a Get. This means that we can attempt to use any
// Git credentials on the host machine to resolve the Get before falling back to
// SSH.
type CustomGitGetter struct {
	*getter.GitGetter
}

// Get overrides the standard Get method transforming SSH urls to their HTTPS
// equivalent. Get then tries to get the new url into the dst, falling back to
// the original SSH url if an HTTPS get fails.
func (g *CustomGitGetter) Get(dst string, u *url.URL) error {
	if u.Scheme != "ssh" {
		return g.GitGetter.Get(dst, u)
	}

	httpsURL, err := TransformSSHToHttps(u)
	if err != nil {
		logging.Logger.Debug().Err(err).Msgf("failed to transform %s to https", u)
		return g.GitGetter.Get(dst, u)
	}

	err = g.GitGetter.Get(dst, httpsURL)
	if err != nil {
		logging.Logger.Debug().Err(err).Msgf("failed to get transformed ssh url %s, retrying with ssh", httpsURL)
		return g.GitGetter.Get(dst, u)
	}

	return nil
}

// IsGitSSHSource returns if the url u is a valid git ssh source. Param u is
// expected to be an url that has been transformed by a go-getter Detect pass,
// removing shorthand and aliased for various sources.
func IsGitSSHSource(u *url.URL) bool {
	if u == nil {
		return false
	}

	if u.Scheme == "ssh" || u.Scheme == "git::ssh" {
		return true
	}

	return false
}

// TransformSSHToHttps transforms a Terraform module source url to an HTTPS
// equivalent. This only handles source urls prefixed with ssh:: or git::ssh. The
// shorthand ssh source referenced here:
// https://developer.hashicorp.com/terraform/language/modules/sources#github e.g.
// "git@github.com:hashicorp/example.git" in not handled by this method as we
// expect the source to already be Detected to the valid longhand equivalent
// before calling this function. This can be achieved by calling
// getter.Detect(src) before calling TransformSSHToHttps.
func TransformSSHToHttps(u *url.URL) (*url.URL, error) {
	if !IsGitSSHSource(u) {
		return u, nil
	}

	hostname := u.Host
	path := strings.TrimPrefix(u.Path, "/")

	// SSH URLs might contain ':' after the host (like 'git@hostname:user/repo.git')
	// We need to replace the first ':' with a '/'
	if idx := strings.Index(path, ":"); idx != -1 {
		path = path[:idx] + "/" + path[idx+1:]
	}

	return &url.URL{
		Scheme: "https",
		Host:   hostname,
		Path:   path,
	}, nil
}
