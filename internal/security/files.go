// Package security contains security related functions, mainly checking a file
// is within a given path, and symlink handling.
package security

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// IsPathAllowed checks if a path is within any of the supplied parent paths.
// It is the containment security boundary: it resolves symlinks anywhere in
// the path (leaf AND intermediate directory symlinks) before comparing against
// the resolved parents, and returns false if the resolved path escapes every
// supplied parent. An in-repo symlink (e.g. evil -> /etc) therefore cannot be
// used to read files outside the allowed parents via a path like evil/passwd
// whose leaf isn't itself a symlink.
//
// Symlink resolution is cached process-wide via resolveCache so repeated calls
// during a parse session don't repeat the (multiple) syscalls per path.
func IsPathAllowed(path string, allowedParents ...string) bool {
	if path == "" {
		path = "."
	}

	pathAbs, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	pathResolved := resolvePathCached(pathAbs)

	for _, parent := range allowedParents {
		if parent == "" {
			continue
		}
		parentAbs, err := filepath.Abs(parent)
		if err != nil {
			continue
		}
		parentResolved := resolvePathCached(parentAbs)

		if strings.HasPrefix(pathResolved+string(filepath.Separator), parentResolved+string(filepath.Separator)) {
			return true
		}
	}

	return false
}

// resolveCache memoises symlink resolution keyed by absolute, uncleaned input
// path. Filesystem state changes during a parse session would invalidate the
// cache, but parser usage doesn't mutate the paths it's inspecting — sources
// are read-only during evaluation.
var resolveCache sync.Map

func resolvePathCached(pathAbs string) string {
	if v, ok := resolveCache.Load(pathAbs); ok {
		return v.(string)
	}
	resolved, err := RecursivelyResolveSymlink(pathAbs)
	if err != nil {
		resolved = pathAbs
	}
	resolved = filepath.Clean(resolved)
	resolveCache.Store(pathAbs, resolved)
	return resolved
}

// RecursivelyResolveSymlink resolves symlinks anywhere in the path, not just
// at the leaf. A leaf-only implementation misses intermediate directory
// symlinks: a legitimate-looking path like /repo-root/dir-link/passwd, where
// dir-link -> /etc, would pass an IsPathAllowed(/repo-root, ...) check because
// the leaf "passwd" isn't a symlink — but reading the path actually opens
// /etc/passwd, which is a path-traversal escape from the allowed-dirs sandbox.
//
// filepath.EvalSymlinks normally errors when any component of the path doesn't
// exist (e.g. existence checks for candidate paths that aren't there yet). For
// those cases we walk up to the longest existing prefix, resolve that, then
// rejoin the missing tail — so prefix-matching still aligns with how parents
// are themselves resolved.
func RecursivelyResolveSymlink(path string) (string, error) {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved, nil
	}
	// path or some ancestor doesn't exist. Walk up until we find an existing
	// ancestor, resolve that, then append the missing tail.
	missing := []string{}
	dir := path
	for {
		parent := filepath.Dir(dir)
		if parent == dir {
			// hit root without finding anything that exists; nothing to
			// resolve, return the original cleaned path.
			return filepath.Clean(path), nil
		}
		missing = append([]string{filepath.Base(dir)}, missing...)
		dir = parent
		resolved, err := filepath.EvalSymlinks(dir)
		if err != nil {
			continue
		}
		return filepath.Join(append([]string{resolved}, missing...)...), nil
	}
}

// IsSymlink returns true if the path's leaf is a symlink.
func IsSymlink(path string) bool {
	fileInfo, err := os.Lstat(path)
	return err == nil && fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink
}
