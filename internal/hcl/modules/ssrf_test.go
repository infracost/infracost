package modules

import (
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsBlockedIP(t *testing.T) {
	tests := []struct {
		ip      string
		blocked bool
	}{
		{"127.0.0.1", true},             // loopback
		{"::1", true},                   // IPv6 loopback
		{"169.254.169.254", true},       // cloud metadata (link-local)
		{"10.0.0.5", true},              // RFC1918
		{"172.16.0.1", true},            // RFC1918
		{"192.168.1.1", true},           // RFC1918
		{"fc00::1", true},               // IPv6 ULA
		{"fe80::1", true},               // IPv6 link-local
		{"0.0.0.0", true},               // unspecified
		{"224.0.0.1", true},             // multicast
		{"::ffff:127.0.0.1", true},      // IPv4-mapped loopback
		{"8.8.8.8", false},              // public
		{"1.1.1.1", false},              // public
		{"2606:4700:4700::1111", false}, // public IPv6
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			require.NotNil(t, ip)
			assert.Equal(t, tt.blocked, IsBlockedIP(ip))
		})
	}

	assert.True(t, IsBlockedIP(nil), "nil IP should fail closed")
}

func TestCheckResolvedIPs(t *testing.T) {
	ips := func(ss ...string) []net.IP {
		out := make([]net.IP, 0, len(ss))
		for _, s := range ss {
			out = append(out, net.ParseIP(s))
		}
		return out
	}

	tests := []struct {
		name    string
		ips     []net.IP
		blocked bool
	}{
		{"no addresses", ips(), false},
		{"all internal", ips("127.0.0.1", "10.0.0.1"), true},
		{"single internal", ips("169.254.169.254"), true},
		{"all public", ips("8.8.8.8", "1.1.1.1"), false},
		// split-horizon: a public record makes it allowed (no false-positive),
		// the dialer still rejects the internal one if it is dialed.
		{"mixed public and internal", ips("8.8.8.8", "127.0.0.1"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkResolvedIPs("example.com", tt.ips)
			if tt.blocked {
				var bae *BlockedAddressError
				assert.ErrorAs(t, err, &bae)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestModuleSourceHost(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		{"https://example.com/mod.zip", "example.com"},
		{"http://127.0.0.1:8099/ssrf", "127.0.0.1"},
		{"git::https://github.com/org/repo?ref=v1.0.0", "github.com"},
		{"git::http://127.0.0.1:8099/ssrf-proof", "127.0.0.1"},
		{"git::ssh://git@10.0.0.5/org/repo", "10.0.0.5"},
		{"git@github.com:org/repo.git", "github.com"},
		{"git@169.254.169.254:org/repo.git", "169.254.169.254"},
		{"hg::http://10.1.2.3/repo", "10.1.2.3"},
		{"github.com/org/repo", "github.com"},
		{"github.com/org/repo//subdir?ref=v1", "github.com"},
		{"./local/module", ""},
		{"../shared", ""},
		{"file:///tmp/module", ""},
	}

	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			assert.Equal(t, tt.want, moduleSourceHost(tt.src))
		})
	}
}

// TestSafeDialer_BlocksLoopback is the core SSRF proof at the transport layer: a
// real server on loopback cannot be reached through a guarded transport, but can
// be when the guard is disabled (the opt-out).
func TestSafeDialer_BlocksLoopback(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	t.Run("guard enabled blocks loopback", func(t *testing.T) {
		client := &http.Client{Transport: GuardTransport(nil, true)}
		resp, err := client.Get(ts.URL)
		if resp != nil {
			_ = resp.Body.Close()
		}
		require.Error(t, err)
		var bae *BlockedAddressError
		assert.ErrorAs(t, err, &bae)
	})

	t.Run("guard disabled allows loopback (opt-out)", func(t *testing.T) {
		client := &http.Client{Transport: GuardTransport(nil, false)}
		resp, err := client.Get(ts.URL)
		require.NoError(t, err)
		_ = resp.Body.Close()
	})
}

// TestPackageFetcher_BlocksInternalSSRF reproduces the FIX-315 PoC through the
// real fetcher entrypoint.
func TestPackageFetcher_BlocksInternalSSRF(t *testing.T) {
	ResetGlobalModuleCache()
	fetcher := NewPackageFetcher(nil, zerolog.Nop())

	srcs := []string{
		"http://127.0.0.1:8099/ssrf-proof-from-untrusted-hcl",
		"http://169.254.169.254/latest/meta-data/iam/security-credentials/ci-role",
		"git::http://127.0.0.1:8100/ssrf",
		"git::ssh://git@10.0.0.5/org/repo",
	}

	for _, src := range srcs {
		t.Run(src, func(t *testing.T) {
			err := fetcher.Fetch(src, filepath.Join(t.TempDir(), "mod"))
			require.Error(t, err)
			var bae *BlockedAddressError
			assert.ErrorAs(t, err, &bae, "expected SSRF guard to block %s", src)
		})
	}
}

// TestPackageFetcher_RedactsCredentialsInError ensures credentials in a blocked
// source never appear in the surfaced error message.
func TestPackageFetcher_RedactsCredentialsInError(t *testing.T) {
	ResetGlobalModuleCache()
	fetcher := NewPackageFetcher(nil, zerolog.Nop())

	for _, src := range []string{
		"https://admin:s3cr3t@127.0.0.1/mod.zip",
		"git::https://admin:s3cr3t@10.0.0.1/org/repo?ref=v1",
		"git::ssh://admin:s3cr3t@192.168.1.5/repo",
	} {
		t.Run(src, func(t *testing.T) {
			err := fetcher.Fetch(src, filepath.Join(t.TempDir(), "mod"))
			require.Error(t, err)
			msg := err.Error()
			assert.NotContains(t, msg, "s3cr3t", "password must not leak")
			assert.NotContains(t, msg, "admin", "username must not leak")
			assert.Contains(t, msg, "****:****@", "credentials should be redacted")
		})
	}
}

// TestPackageFetcher_AllowPrivateSourcesOptOut shows the opt-out disables the guard.
func TestPackageFetcher_AllowPrivateSourcesOptOut(t *testing.T) {
	ResetGlobalModuleCache()
	fetcher := NewPackageFetcher(nil, zerolog.Nop(), WithAllowPrivateSources(true))

	err := fetcher.Fetch("git::http://127.0.0.1:1/ssrf", filepath.Join(t.TempDir(), "mod"))
	require.Error(t, err) // nothing is listening, so it still fails...
	var bae *BlockedAddressError
	assert.NotErrorAs(t, err, &bae, "guard should be disabled by the opt-out")
}
