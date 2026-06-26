package template

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupSymlinkRepo builds a repo tree that mirrors the FIX-313 PoC:
//
//	repo/                 -- the repository root the parser is confined to
//	  inside/secret.tf    -- a legitimate in-repo file
//	  evil -> outside     -- an intermediate symlink escaping the repo
//	  finlink -> outside/secret  -- a leaf symlink escaping the repo
//	outside/secret        -- a secret that lives outside the repo root
//
// It returns the repo dir.
func setupSymlinkRepo(t *testing.T) string {
	t.Helper()

	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	outside := filepath.Join(tmp, "outside")

	require.NoError(t, os.MkdirAll(filepath.Join(repo, "inside"), 0700))
	require.NoError(t, os.MkdirAll(outside, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(repo, "inside", "secret.tf"), []byte("in-repo"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(outside, "secret"), []byte("TOP-SECRET"), 0600))
	require.NoError(t, os.Symlink(outside, filepath.Join(repo, "evil")))
	require.NoError(t, os.Symlink(filepath.Join(outside, "secret"), filepath.Join(repo, "finlink")))

	return repo
}

func TestParser_readFile_confinement(t *testing.T) {
	repo := setupSymlinkRepo(t)
	p := NewParser(repo, Variables{}, nil)

	// Legitimate in-repo read works.
	assert.Equal(t, "in-repo", p.readFile("inside/secret.tf"))

	// Lexical parent traversal is blocked.
	assert.Panics(t, func() { p.readFile("../outside/secret") },
		"parent-directory traversal must be blocked")

	// Leaf symlink escaping the repo is blocked.
	assert.Panics(t, func() { p.readFile("finlink") },
		"leaf symlink escaping the repo must be blocked")

	// Intermediate symlink escaping the repo is blocked (the FIX-313 gap).
	assert.Panics(t, func() { p.readFile("evil/secret") },
		"intermediate symlink escaping the repo must be blocked")
}

func TestParser_pathExists_confinement(t *testing.T) {
	repo := setupSymlinkRepo(t)
	p := NewParser(repo, Variables{}, nil)

	// Legitimate in-repo path exists.
	assert.True(t, p.pathExists("inside", "secret.tf"))

	// Parent traversal to a real file outside the repo reports false.
	assert.False(t, p.pathExists(".", "../outside/secret"),
		"parent-directory traversal must not be reported as existing")

	// Intermediate symlink escaping the repo reports false even though the
	// underlying file exists.
	assert.False(t, p.pathExists("evil", "secret"),
		"path traversing an intermediate symlink must not be reported as existing")

	// A base that is itself an escaping symlink reports false.
	assert.False(t, p.pathExists("evil", ""),
		"a base that escapes the repo via a symlink must not be reported as existing")
}

func TestParser_isDir_confinement(t *testing.T) {
	repo := setupSymlinkRepo(t)
	p := NewParser(repo, Variables{}, nil)

	// Legitimate in-repo directory.
	assert.True(t, p.isDir("inside"))

	// The intermediate symlink resolves to an out-of-repo directory and must
	// not be reported as an in-repo directory.
	assert.False(t, p.isDir("evil"),
		"a symlinked directory escaping the repo must not be reported as a dir")
}

func TestParser_matchPaths_confinement(t *testing.T) {
	repo := setupSymlinkRepo(t)
	p := NewParser(repo, Variables{}, nil)

	matches := p.matchPaths(":dir/secret")

	// Nothing reached via the escaping `evil` symlink may appear.
	for _, m := range matches {
		assert.NotEqual(t, "evil", m["_dir"],
			"matchPaths must not return entries reached via an escaping symlink")
	}
}

// TestParser_Compile_readFile_blocksIntermediateSymlink drives the real
// template entrypoint (Compile) with the exact PoC payload from FIX-313 and
// asserts the out-of-repo secret is never rendered into the output.
func TestParser_Compile_readFile_blocksIntermediateSymlink(t *testing.T) {
	repo := setupSymlinkRepo(t)
	p := NewParser(repo, Variables{}, nil)

	var buf bytes.Buffer
	err := p.Compile(`{{ readFile "evil/secret" }}`, &buf)

	require.Error(t, err, "rendering a template that escapes the repo must error")
	assert.NotContains(t, buf.String(), "TOP-SECRET",
		"the out-of-repo secret must never be rendered into the output")
}
