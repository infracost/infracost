package modules

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	getter "github.com/hashicorp/go-getter"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransformSSHToHTTPS(t *testing.T) {
	testCases := []struct {
		sshURL   *url.URL
		expected string
	}{
		{&url.URL{Scheme: "ssh", User: url.User("git"), Path: "user/repo.git", Host: "github.com"}, "https://github.com/user/repo.git"},
		{&url.URL{Scheme: "https", Path: "user/repo.git", Host: "github.com"}, "https://github.com/user/repo.git"},
		{&url.URL{Scheme: "git::ssh", User: url.User("git"), Path: "user/repo.git", Host: "github.com"}, "https://github.com/user/repo.git"},
	}

	for _, tc := range testCases {
		transformed, err := TransformSSHToHttps(tc.sshURL)
		assert.NoError(t, err)

		assert.Equal(t, tc.expected, transformed.String())
	}
}

func TestPackageFetcher_fetch_RemoteCache(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	logger := zerolog.New(io.Discard)
	mock := &mockRemoteCache{
		cache: make(map[string]mockCacheEntry),
	}

	tests := []struct {
		name          string
		setupCache    func(*mockRemoteCache, string)
		expectedCalls map[string]int
		expectedError bool
	}{
		{
			name: "should use cached module from remote cache",
			setupCache: func(c *mockRemoteCache, tmpDir string) {
				// Create a mock module directory with content
				moduleDir := filepath.Join(tmpDir, "cached_module")
				require.NoError(t, os.MkdirAll(moduleDir, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(moduleDir, "main.tf"), []byte(`
					variable "instance_type" {
						type = string
						default = "t3.micro"
					}
				`), 0600))
				// Store the module directory path in the cache
				require.NoError(t, c.Put("git::https://github.com/terraform-aws-modules/terraform-aws-vpc?ref=v5.15.0", moduleDir, 24*time.Hour))
			},
			expectedCalls: map[string]int{
				"Exists": 1,
				"Get":    1,
				"Put":    0,
			},
		},
		{
			name:       "should cache module to remote cache when not found",
			setupCache: func(c *mockRemoteCache, tmpDir string) {},
			expectedCalls: map[string]int{
				"Exists": 1,
				"Get":    0,
				"Put":    1,
			},
		},
		{
			name: "should handle remote cache errors gracefully",
			setupCache: func(c *mockRemoteCache, tmpDir string) {
				c.shouldError = true
			},
			expectedCalls: map[string]int{
				"Exists": 1,
				"Get":    0,
				"Put":    1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			mock.ResetCache()
			tt.setupCache(mock, tmpDir)
			mock.ResetCounters()

			// Create the test Terraform configuration
			err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(`
				module "ec2_instance" {
					source  = "git::https://github.com/terraform-aws-modules/terraform-aws-vpc?ref=v5.15.0"
				}
			`), 0600)
			require.NoError(t, err)

			fetcher := NewPackageFetcher(mock, logger)

			err = fetcher.Fetch("git::https://github.com/terraform-aws-modules/terraform-aws-vpc?ref=v5.15.0", filepath.Join(tmpDir, "module"))
			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedCalls["Exists"], mock.existsCalls, "wrong number of Exists calls")
			assert.Equal(t, tt.expectedCalls["Get"], mock.getCalls, "wrong number of Get calls")
			assert.Equal(t, tt.expectedCalls["Put"], mock.putCalls, "wrong number of Put calls")
		})
	}
}

type mockCacheEntry struct {
	path      string
	expiresAt time.Time
}

type mockRemoteCache struct {
	cache       map[string]mockCacheEntry
	existsCalls int
	getCalls    int
	putCalls    int
	shouldError bool
	mu          sync.Mutex
}

func (m *mockRemoteCache) Exists(key string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.existsCalls++

	if m.shouldError {
		return false, fmt.Errorf("mock remote cache error")
	}

	entry, exists := m.cache[key]

	// Check if the entry has expired
	if time.Now().After(entry.expiresAt) {
		delete(m.cache, key)
		return false, nil
	}

	return exists, nil
}

func (m *mockRemoteCache) Get(key string, dest string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getCalls++

	if m.shouldError {
		return fmt.Errorf("mock remote cache error")
	}

	entry, exists := m.cache[key]
	if !exists {
		return fmt.Errorf("key not found: %s", key)
	}

	// Use go-getter to copy the directory
	client := &getter.Client{
		Src:  fmt.Sprintf("file://%s", entry.path),
		Dst:  dest,
		Mode: getter.ClientModeDir,
	}

	return client.Get()
}

func (m *mockRemoteCache) Put(key string, src string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.putCalls++

	if m.shouldError {
		return fmt.Errorf("mock remote cache error")
	}

	m.cache[key] = mockCacheEntry{
		path:      src,
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}

func (m *mockRemoteCache) ResetCache() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cache = make(map[string]mockCacheEntry)
}

func (m *mockRemoteCache) ResetCounters() {
	m.existsCalls = 0
	m.getCalls = 0
	m.putCalls = 0
	m.shouldError = false
}
