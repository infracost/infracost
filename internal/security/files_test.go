package security

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIsPathAllowed_BlocksIntermediateSymlinkTraversal locks in the security
// posture that intermediate directory symlinks are resolved before the
// allowed-parents check. Before this fix, a legitimately-named path like
// /repo-root/dir-link/passwd where dir-link points outside the repo would
// pass IsPathAllowed because the leaf wasn't a symlink — letting parser/
// template code read files outside the sandbox.
func TestIsPathAllowed_BlocksIntermediateSymlinkTraversal(t *testing.T) {
	// repo/    -- the allowed parent
	//   inside/ -- a directory inside repo (touched so it exists)
	//   link  -- symlink pointing OUTSIDE repo to ../outside
	// outside/
	//   secret -- the file we shouldn't be able to reach via repo/link/secret
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	outside := filepath.Join(tmp, "outside")
	require.NoError(t, os.MkdirAll(filepath.Join(repo, "inside"), 0700))
	require.NoError(t, os.MkdirAll(outside, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(outside, "secret"), []byte("nope"), 0600))
	require.NoError(t, os.Symlink(outside, filepath.Join(repo, "link")))

	// /repo/inside/file (entirely within repo) — allowed
	assert.True(t, IsPathAllowed(filepath.Join(repo, "inside", "anything"), repo),
		"a path inside the allowed parent should be allowed")

	// /repo/link/secret — link points OUTSIDE repo. The leaf "secret" is
	// not a symlink, but the intermediate "link" is. This is the case the
	// old leaf-only resolver missed.
	assert.False(t, IsPathAllowed(filepath.Join(repo, "link", "secret"), repo),
		"a path traversing an intermediate symlink that escapes the allowed parent must be rejected")

	// /repo/link (the symlink itself, leaf is the symlink) — also points
	// outside, must be rejected.
	assert.False(t, IsPathAllowed(filepath.Join(repo, "link"), repo),
		"a leaf-symlink that targets outside the allowed parent must be rejected")
}

// TestIsPathAllowed_HandlesNonExistentPaths verifies the longest-existing-
// prefix resolver — existence checks walk candidate paths that don't exist
// yet, and the security check must still recognize them as inside the allowed
// parent.
func TestIsPathAllowed_HandlesNonExistentPaths(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	require.NoError(t, os.MkdirAll(repo, 0700))

	// Candidate file doesn't exist yet. Still inside repo.
	assert.True(t, IsPathAllowed(filepath.Join(repo, "doesnotexist.tf"), repo),
		"a non-existent path inside the allowed parent should be allowed")

	// Candidate file with a non-existent intermediate dir, still inside repo.
	assert.True(t, IsPathAllowed(filepath.Join(repo, "missing", "nested", "file.tf"), repo),
		"a deeply nested non-existent path inside the allowed parent should be allowed")

	// Candidate file outside the repo — even if it doesn't exist, must be rejected.
	assert.False(t, IsPathAllowed(filepath.Join(tmp, "elsewhere", "file.tf"), repo),
		"a non-existent path outside the allowed parent must be rejected")
}
