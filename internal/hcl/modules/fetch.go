package modules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	getter "github.com/hashicorp/go-getter"
	"github.com/otiai10/copy"
	"github.com/sirupsen/logrus"
)

// PackageFetcher downloads modules from a remote source to the given destination
// This supports all the non-local and non-Terraform registry sources listed here: https://www.terraform.io/language/modules/sources
type PackageFetcher struct {
	cache  map[string]string
	logger *logrus.Entry
}

// NewPackageFetcher constructs a new package fetcher
func NewPackageFetcher(logger *logrus.Entry) *PackageFetcher {
	return &PackageFetcher{
		cache:  make(map[string]string),
		logger: logger,
	}
}

// fetch downloads the remote module using the go-getter library
// See: https://github.com/hashicorp/go-getter
func (r *PackageFetcher) fetch(moduleAddr string, dest string) error {
	if prevDest, ok := r.cache[moduleAddr]; ok {
		r.logger.Debugf("module %s already downloaded, copying from '%s' to '%s'", moduleAddr, prevDest, dest)

		err := os.Mkdir(dest, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create directory '%s': %w", dest, err)
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
			return fmt.Errorf("failed to copy module from '%s' to '%s': %w", prevDest, dest, err)
		}

		return nil
	}
	var cached []string
	for k := range r.cache {
		cached = append(cached, k)
	}

	r.logger.WithFields(logrus.Fields{"cached_module_addresses": cached}).Debugf("module %s does not exist in cache, proceeding to download", moduleAddr)

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
		Mode:          getter.ClientModeDir,
		Decompressors: decompressors,
		// We don't need to specify any of the Getters, since Terraform uses the same as the default Getter values,
		// but if we do need to at some point we can specify them here:
		// Getters: getters,
	}

	r.cache[moduleAddr] = dest

	err := client.Get()
	if err != nil {
		return fmt.Errorf("could not download module %s to cache %w", moduleAddr, err)
	}

	return nil
}
