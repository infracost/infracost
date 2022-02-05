package modules

import (
	getter "github.com/hashicorp/go-getter"
)

// RemoteLoader is a loader that downloads modules from a remote source to the given destination
// This supports all the non-local and non-Terraform registry sources listed here: https://www.terraform.io/language/modules/sources
type RemoteLoader struct {
	dest string
}

// NewRemoteLoader constructs a new remote loader
func NewRemoteLoader(dest string) *RemoteLoader {
	return &RemoteLoader{dest: dest}
}

// downloadModule downloads the remote module
func (r *RemoteLoader) downloadModule(moduleAddr string) error {
	return downloadRemoteModule(moduleAddr, r.dest)
}

// downloadRemoteModule downloads the remote module using the go-getter library
// See: https://github.com/hashicorp/go-getter
func downloadRemoteModule(source string, dest string) error {
	decompressors := map[string]getter.Decompressor{}
	for k, decompressor := range getter.Decompressors {
		decompressors[k] = decompressor
	}
	// This one is added by Terraform here: https://github.com/hashicorp/terraform/blob/affe2c329561f40f13c0e94f4570321977527a77/internal/getmodules/getter.go#L64
	// But is not in the list of default compressors here: https://github.com/hashicorp/go-getter/blob/main/decompress.go#L32
	// I'm not sure if we really need it, but added it just in case/
	decompressors["tar.tbz2"] = new(getter.TarBzip2Decompressor)

	client := getter.Client{
		Src:           source,
		Dst:           dest,
		Pwd:           dest,
		Mode:          getter.ClientModeDir,
		Decompressors: decompressors,
		// We don't need to specify any of the Getters, since Terraform uses the same as the default Getter values,
		// but if we do need to at some point we can specify them here:
		// Getters: getters,
	}

	return client.Get()
}
