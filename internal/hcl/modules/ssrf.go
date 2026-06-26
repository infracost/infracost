package modules

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"syscall"
	"time"

	"github.com/hashicorp/go-getter"
)

// This file implements an SSRF guard for the outbound requests Infracost makes
// when fetching Terraform modules. When Infracost runs in CI against an
// untrusted pull request, the PR author controls every `module { source = ... }`
// block, and each source is fetched without any host allowlist. Without a guard
// that lets an attacker point Infracost at loopback, link-local (cloud
// metadata, e.g. 169.254.169.254) or other private addresses that are only
// reachable from inside the CI network. See FIX-315.
//
// There are two layers:
//
//   - SafeDialer / GuardTransport guard the HTTP transport used by the http(s)
//     getter and the registry client. The check runs at connect time on the
//     resolved IP, so it catches both "evil.com resolves to 127.0.0.1" (DNS
//     rebinding) and "a public host 301-redirects to an internal address" (each
//     redirect re-dials and is re-checked).
//
//   - CheckHost is a pre-fetch check for getters that shell out (git, hg) or use
//     their own SDK clients (s3, gcs) and therefore never touch our transport.
//
// Connections routed through an operator-configured registry proxy
// (INFRACOST_REGISTRY_PROXY) are exempt from the guard: the proxy is trusted
// infrastructure and may legitimately live on, or forward to, an internal
// address.

// IsBlockedIP reports whether ip is an address that untrusted module sources
// must not be allowed to reach: loopback, RFC1918 private ranges, IPv6 ULA,
// link-local (which includes the 169.254.169.254 cloud metadata endpoint),
// multicast, and the unspecified address. A nil IP is treated as blocked so the
// guard fails closed.
func IsBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}

	// Normalize IPv4-in-IPv6 so the net.IP helpers below classify it as IPv4.
	if v4 := ip.To4(); v4 != nil {
		ip = v4
	}

	return ip.IsLoopback() || // 127.0.0.0/8, ::1
		ip.IsPrivate() || // 10/8, 172.16/12, 192.168/16, fc00::/7 (ULA)
		ip.IsLinkLocalUnicast() || // 169.254.0.0/16 (metadata), fe80::/10
		ip.IsLinkLocalMulticast() ||
		ip.IsInterfaceLocalMulticast() ||
		ip.IsMulticast() ||
		ip.IsUnspecified() // 0.0.0.0, ::
}

// BlockedAddressError is returned when the SSRF guard rejects a target. The
// message is deliberately terse so it does not leak internal-network details.
type BlockedAddressError struct {
	Host string
	IP   net.IP
}

func (e *BlockedAddressError) Error() string {
	if e.Host != "" && (e.IP == nil || e.Host != e.IP.String()) {
		return fmt.Sprintf("refusing to fetch module from %s (resolves to internal/loopback address %s)", e.Host, e.IP)
	}

	return fmt.Sprintf("refusing to fetch module from internal/loopback address %s", e.IP)
}

// CheckHost resolves host (a hostname or IP literal) and returns a
// BlockedAddressError if it - or every address it resolves to - is internal.
//
// It is the pre-fetch counterpart to the dialer guard, for getters that perform
// their own DNS resolution and connection (git, hg, s3, gcs). A hostname that
// also resolves to at least one public address is allowed: for http(s) the
// dialer pins a safe address at connect time, and for the shell-out getters this
// avoids false-positives on split-horizon DNS (where a host has both a public
// and an internal record). A DNS failure returns nil: there is nothing to safely
// block and the fetch will fail on its own if the source is invalid.
func CheckHost(ctx context.Context, host string) error {
	if host == "" {
		return nil
	}

	if ip := net.ParseIP(host); ip != nil {
		if IsBlockedIP(ip) {
			return &BlockedAddressError{Host: host, IP: ip}
		}
		return nil
	}

	addrs, err := (&net.Resolver{}).LookupIPAddr(ctx, host)
	if err != nil {
		return nil
	}

	ips := make([]net.IP, 0, len(addrs))
	for _, a := range addrs {
		ips = append(ips, a.IP)
	}

	return checkResolvedIPs(host, ips)
}

// checkResolvedIPs blocks only when every resolved address is internal. If the
// host resolves to no addresses, there is nothing to block.
func checkResolvedIPs(host string, ips []net.IP) error {
	if len(ips) == 0 {
		return nil
	}

	var firstBlocked net.IP
	for _, ip := range ips {
		if !IsBlockedIP(ip) {
			// At least one public address: let the connection proceed (the
			// dialer guard will still reject any internal address it is asked to
			// connect to).
			return nil
		}
		if firstBlocked == nil {
			firstBlocked = ip
		}
	}

	return &BlockedAddressError{Host: host, IP: firstBlocked}
}

// SafeDialer is a net dialer that rejects connections to internal addresses
// (see IsBlockedIP). The check runs in the dialer's Control hook, which fires
// after DNS resolution but before the socket connects, with the concrete IP
// that is about to be dialed - this is what gives us DNS-rebinding and redirect
// safety for free. When enabled is false the guard is a no-op.
type SafeDialer struct {
	plain   *net.Dialer
	guarded *net.Dialer
	enabled bool
}

// NewSafeDialer builds a SafeDialer.
func NewSafeDialer(enabled bool) *SafeDialer {
	base := func() *net.Dialer {
		return &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}
	}

	guarded := base()
	guarded.Control = func(_, address string, _ syscall.RawConn) error {
		host, _, err := net.SplitHostPort(address)
		if err != nil {
			host = address
		}

		ip := net.ParseIP(host)
		if ip == nil {
			// Control is always called with a resolved IP:port, so a parse
			// failure is unexpected. Fail closed rather than risk a bypass.
			return fmt.Errorf("ssrf guard: could not parse dial address %q", address)
		}

		if IsBlockedIP(ip) {
			return &BlockedAddressError{Host: host, IP: ip}
		}

		return nil
	}

	return &SafeDialer{
		plain:   base(),
		guarded: guarded,
		enabled: enabled,
	}
}

// DialContext dials addr, applying the SSRF guard unless the guard is disabled.
func (d *SafeDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	if !d.enabled {
		return d.plain.DialContext(ctx, network, addr)
	}

	return d.guarded.DialContext(ctx, network, addr)
}

// GuardTransport returns a clone of base with the SSRF guard installed on its
// dialer. base is not mutated. If base is nil, http.DefaultTransport is used as
// the template.
func GuardTransport(base *http.Transport, enabled bool) *http.Transport {
	if base == nil {
		if dt, ok := http.DefaultTransport.(*http.Transport); ok {
			base = dt
		} else {
			base = &http.Transport{}
		}
	}

	t := base.Clone()
	t.DialContext = NewSafeDialer(enabled).DialContext

	return t
}

// hostIsProxied reports whether host matches one of the configured proxy hosts
// (exact match or subdomain), mirroring ConditionalTransport's routing. Such
// hosts are exempt from the pre-fetch check because the connection goes to the
// trusted proxy rather than directly to host.
func hostIsProxied(host string, proxyHosts []string) bool {
	for _, ph := range proxyHosts {
		if host == ph || strings.HasSuffix(host, "."+ph) {
			return true
		}
	}
	return false
}

// moduleSourceHost extracts the network host from a go-getter module source. It
// uses go-getter's own detectors to normalize the many shorthand and aliased
// forms - "github.com/org/repo", "git@host:org/repo", "git::ssh://...",
// "hg::http://...", "s3::https://..." - into a canonical URL, then parses out the
// host. It returns "" when there is no network host to check (local paths,
// file://, unparseable sources).
func moduleSourceHost(src string) string {
	detected, err := getter.Detect(strings.TrimSpace(src), "", getter.Detectors)
	if err != nil || detected == "" {
		detected = src
	}

	// Strip a go-getter forced-getter prefix such as "git::" or "hg::" (distinct
	// from a URL scheme, which uses "://").
	if i := strings.Index(detected, "::"); i > 0 && !strings.ContainsAny(detected[:i], "/:@") {
		detected = detected[i+2:]
	}

	// Drop the go-getter subdir ("//path") and query ("?ref=") suffixes.
	if i := strings.Index(detected, "?"); i >= 0 {
		detected = detected[:i]
	}

	u, err := url.Parse(detected)
	if err != nil {
		return ""
	}

	switch u.Scheme {
	case "", "file":
		return ""
	}

	return u.Hostname()
}
