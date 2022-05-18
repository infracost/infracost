package modules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	getter "github.com/hashicorp/go-getter"
	"github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"
)

// PackageFetcher downloads modules from a remote source to the given destination
// This supports all the non-local and non-Terraform registry sources listed here: https://www.terraform.io/language/modules/sources
type PackageFetcher struct {
	cache map[string]string
}

// NewPackageFetcher constructs a new package fetcher
func NewPackageFetcher() *PackageFetcher {
	return &PackageFetcher{
		cache: make(map[string]string),
	}
}

// fetch downloads the remote module using the go-getter library
// See: https://github.com/hashicorp/go-getter
func (r *PackageFetcher) fetch(moduleAddr string, dest string) error {
	err := os.MkdirAll(dest, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Failed to create directory '%s': %w", dest, err)
	}

	if prevDest, ok := r.cache[moduleAddr]; ok {
		log.Debugf("Module %s already downloaded, copying from '%s' to '%s'", moduleAddr, prevDest, dest)

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
			return fmt.Errorf("Failed to copy module from '%s' to '%s': %w", prevDest, dest, err)
		}

		return nil
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
		Mode:          getter.ClientModeDir,
		Decompressors: decompressors,
		// We don't need to specify any of the Getters, since Terraform uses the same as the default Getter values,
		// but if we do need to at some point we can specify them here:
		// Getters: getters,
	}

	r.cache[moduleAddr] = dest

	return client.Get()
}
