package modules

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/infracost/infracost/internal/util"

	"github.com/hashicorp/go-getter"
	"github.com/otiai10/copy"
	"github.com/rs/zerolog"

	"github.com/infracost/infracost/internal/logging"
)

var tagRegex = regexp.MustCompile(`^v?\d+\.\d+\.\d+`)
var commitRegex = regexp.MustCompile(`^([0-9a-f]{40})|([0-9a-f]{7})$`)
var defaultTTL = 24 * time.Hour
var tagCommitTTL = 30 * 24 * time.Hour

const defaultModuleRetrieveTimeout = 3 * time.Minute

// PackageFetcher downloads modules from a remote source to the given destination
// This supports all the non-local and non-Terraform registry sources listed here: https://www.terraform.io/language/modules/sources
type PackageFetcher struct {
	remoteCache         RemoteCache
	logger              zerolog.Logger
	getters             map[string]getter.Getter
	publicModuleChecker PublicModuleChecker
}

// use a global cache to avoid downloading the same module multiple times for each project
var localCache sync.Map
var errorCache sync.Map

func ResetGlobalModuleCache() {
	localCache = sync.Map{}
	errorCache = sync.Map{}
}

type PackageFetcherOpts func(*PackageFetcher)

type ConditionalTransport struct {
	ProxyHosts []string
	ProxyURL   string
	Inner      http.RoundTripper
}

func (t *ConditionalTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for _, host := range t.ProxyHosts {
		if req.URL.Host == host || strings.HasSuffix(req.URL.Host, "."+host) {
			if parsed, err := url.Parse(t.ProxyURL); err == nil {
				proxy := http.ProxyURL(parsed)
				proxyTransport := &http.Transport{Proxy: proxy}
				return proxyTransport.RoundTrip(req)
			}
		}
	}
	if t.Inner != nil {
		return t.Inner.RoundTrip(req)
	}
	return http.DefaultTransport.RoundTrip(req)
}

// NewPackageFetcher constructs a new package fetcher
func NewPackageFetcher(remoteCache RemoteCache, logger zerolog.Logger, opts ...PackageFetcherOpts) *PackageFetcher {
	getters := make(map[string]getter.Getter, len(getter.Getters))
	for k, g := range getter.Getters {
		getters[k] = g
	}
	getters["git"] = &CustomGitGetter{
		&getter.GitGetter{
			Timeout: defaultModuleRetrieveTimeout,
		},
	}
	if proxy := os.Getenv("INFRACOST_REGISTRY_PROXY"); proxy != "" {
		httpGetter := &getter.HttpGetter{
			Netrc: true,
			Client: &http.Client{
				Transport: &ConditionalTransport{
					ProxyHosts: []string{"terraform.io"},
					ProxyURL:   proxy,
					Inner:      http.DefaultTransport,
				},
			},
		}
		getters["http"] = httpGetter
		getters["https"] = httpGetter
	}

	p := &PackageFetcher{
		remoteCache: remoteCache,
		logger:      logger,
		getters:     getters,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func WithGetters(getters map[string]getter.Getter) PackageFetcherOpts {
	return func(p *PackageFetcher) {
		for k, g := range getters {
			p.getters[k] = g
		}
	}
}

// Option to provide a checker to determine if a module is public
func WithPublicModuleChecker(publicModuleChecker PublicModuleChecker) PackageFetcherOpts {
	return func(p *PackageFetcher) {
		p.publicModuleChecker = publicModuleChecker
	}
}

// fetch downloads the remote module using the go-getter library
// See: https://github.com/hashicorp/go-getter
func (p *PackageFetcher) Fetch(moduleAddr string, dest string) error {
	if strings.HasPrefix(moduleAddr, "file://") {
		// Skip to the remote getter so it just copies this instead of
		// looking up the cache
		_, err := p.fetchFromRemote(moduleAddr, dest)
		if err != nil {
			return fmt.Errorf("error fetching file:// module %s from remote: %w", moduleAddr, err)
		}
		return nil
	}

	fetched, err := p.fetchFromLocalCache(moduleAddr, dest)
	if fetched {
		p.logger.Trace().Msgf("cache hit (local): %s", util.RedactUrl(moduleAddr))
		return nil
	}

	if err != nil {
		// Log the error at debug level because the local cache might be invalid and in that case we just refetch it so it's not a big deal
		p.logger.Debug().Msgf("error fetching module %s from local cache: %s", util.RedactUrl(moduleAddr), err)
	}

	var isPublicModule bool
	fetched, isPublicModule, err = p.fetchFromRemoteCache(moduleAddr, dest)
	if fetched {
		p.logger.Trace().Msgf("cache hit (remote): %s", moduleAddr)
		localCache.Store(moduleAddr, dest)
		return nil
	}

	if err != nil {
		// Log the error at debug level because the remote cache might be invalid and in that case we just refetch it so it's not a big deal
		p.logger.Debug().Msgf("error fetching module %s from remote cache: %s", util.RedactUrl(moduleAddr), err)
	}

	p.logger.Trace().Msgf("cache miss: %s", moduleAddr)

	_, err = p.fetchFromRemote(moduleAddr, dest)
	if err != nil {
		return fmt.Errorf("error fetching module %s from remote: %w", util.RedactUrl(moduleAddr), err)
	}

	localCache.Store(moduleAddr, dest)

	if p.remoteCache != nil {
		ttl := determineTTL(moduleAddr)
		p.logger.Debug().Msgf("putting module %s into remote cache with ttl %s", util.RedactUrl(moduleAddr), ttl)
		err = p.remoteCache.Put(moduleAddr, dest, ttl, isPublicModule)
		if err != nil {
			p.logger.Warn().Msgf("error putting module %s into remote cache: %s", util.RedactUrl(moduleAddr), err)
		}
	}

	return nil
}

func (p *PackageFetcher) isPublicModule(moduleAddr string) bool {
	if p.publicModuleChecker == nil {
		return false
	}
	result, err := p.publicModuleChecker.IsPublicModule(moduleAddr)
	if err != nil {
		p.logger.Debug().Msgf("Failed to check if %s is a public module: %v", util.RedactUrl(moduleAddr), err)
	}
	return result
}

func (p *PackageFetcher) fetchFromLocalCache(moduleAddr, dest string) (bool, error) {
	v, ok := localCache.Load(moduleAddr)
	if !ok {
		return false, nil
	}

	prevDest, _ := v.(string)

	if prevDest == dest {
		return true, nil
	}

	p.logger.Debug().Msgf("module %s already downloaded, copying from '%s' to '%s'", util.RedactUrl(moduleAddr), prevDest, dest)

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

func (p *PackageFetcher) fetchFromRemoteCache(moduleAddr, dest string) (bool, bool, error) {
	if p.remoteCache == nil {
		return false, false, nil
	}

	// check if the module is public by HEADing the module address
	public := p.isPublicModule(moduleAddr)

	ok, err := p.remoteCache.Exists(moduleAddr, public)
	if err != nil {
		return false, public, err
	}

	if !ok {
		return false, public, nil
	}

	err = p.remoteCache.Get(moduleAddr, dest, public)
	if err != nil {
		return false, public, err
	}

	return true, public, nil
}

func (p *PackageFetcher) fetchFromRemote(moduleAddr, dest string) (bool, error) {

	if err, ok := errorCache.Load(moduleAddr); ok {
		return false, err.(error)
	}

	decompressors := map[string]getter.Decompressor{}
	for k, decompressor := range getter.Decompressors {
		decompressors[k] = decompressor
	}
	// This one is added by Terraform here: https://github.com/hashicorp/terraform/blob/affe2c329561f40f13c0e94f4570321977527a77/internal/getmodules/getter.go#L64
	// But is not in the list of default compressors here: https://github.com/hashicorp/go-getter/blob/main/decompress.go#L32
	// I'm not sure if we really need it, but added it just in case/
	decompressors["tar.tbz2"] = new(getter.TarBzip2Decompressor)

	client := getter.Client{
		Src:           moduleAddr,
		Dst:           dest,
		Pwd:           dest,
		Mode:          getter.ClientModeAny,
		Decompressors: decompressors,
		Getters:       p.getters,
	}

	err := client.Get()
	if err != nil {
		errorCache.Store(moduleAddr, err)
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

// CustomGitGetter extends the standard GitGetter and normalizes SSH and HTTPS URLs
// so it can attempt to use the Git credentials on the host machine to resolve the
// Get before falling back to the original method.
// SSH URLs are transformed to their HTTPS equivalent before attempting a Get.
// HTTPS URLS are stripped of any credentials.
type CustomGitGetter struct {
	*getter.GitGetter
}

// Get overrides the standard Get method to normalize SSH and HTTPS URLs before
// attempting a Get. If the normalized URL fails it falls back to the original
// URL.
func (g *CustomGitGetter) Get(dst string, u *url.URL) error {
	httpsURL, err := NormalizeGitURLToHTTPS(u)
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

	if u.Scheme == "ssh" || u.Scheme == "git::ssh" || u.Scheme == "git+ssh" {
		return true
	}

	return false
}

// NormalizeGitURLToHTTPS normalizes a Git source url to an HTTPS equivalent.
// It supports SSH URLs, as well as HTTPS URLS with usernames.
// This allows us to convert Terraform module source urls to HTTPS URLs, so we can
// attempt to download over HTTPS first using the existing credentials we have.
// It can also be used for generating links to resources within the module.
// There is a special case for Azure DevOps SSH URLs to handle converting them
// to the equivalent HTTPS URL.
func NormalizeGitURLToHTTPS(u *url.URL) (*url.URL, error) {
	hostname := u.Host
	// Strip the port if it's an SSH url
	if IsGitSSHSource(u) {
		hostname = strings.Split(hostname, ":")[0]
	}

	path := strings.TrimPrefix(u.Path, "/")

	// Handle Azure DevOps SSH URLs
	if hostname == "ssh.dev.azure.com" {
		// Azure DevOps URLs need special handling
		// Convert from: ssh://git@ssh.dev.azure.com/v3/org/project/repo
		// To: https://dev.azure.com/org/project/_git/repo
		parts := strings.Split(path, "/")
		if len(parts) >= 4 && strings.HasPrefix(parts[0], "v") {
			org := parts[1]
			project := parts[2]
			repo := parts[3]
			return &url.URL{
				Scheme: "https",
				Host:   "dev.azure.com",
				Path:   fmt.Sprintf("/%s/%s/_git/%s", org, project, repo),
			}, nil
		}
		return nil, fmt.Errorf("invalid Azure DevOps SSH URL format")
	}

	// Strip .git from the end of the path
	path = strings.TrimSuffix(path, ".git")

	return &url.URL{
		Scheme:      "https",
		Host:        hostname,
		Path:        path,
		RawQuery:    u.RawQuery,
		RawFragment: u.RawFragment,
	}, nil
}
