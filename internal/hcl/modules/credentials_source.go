package modules

import (
	"net/http"

	svchost "github.com/hashicorp/terraform-svchost"
	"github.com/hashicorp/terraform-svchost/auth"
)

type HostCredentials struct {
	token string
}

func (c HostCredentials) Token() string { return c.token }

func (c HostCredentials) PrepareRequest(req *http.Request) {
	if c.token == "" {
		return
	}

	req.Header.Set("Authorization", "Bearer "+c.Token())
}

type FindTokenForHost func(host string) string

type CredentialsSource struct {
	findTokenForHost FindTokenForHost
}

func NewCredentialsSource(findTokenForHost FindTokenForHost) *CredentialsSource {
	return &CredentialsSource{
		findTokenForHost: findTokenForHost,
	}
}

// ForHost returns a non-nil HostCredentials if the source has credentials
// available for the host, and a nil HostCredentials if it does not.
func (s *CredentialsSource) ForHost(host svchost.Hostname) (auth.HostCredentials, error) {
	return HostCredentials{token: s.findTokenForHost(host.ForDisplay())}, nil
}

// StoreForHost is unimplemented but is required for the auth.CredentialsSource interface.
func (s *CredentialsSource) StoreForHost(host svchost.Hostname, credentials auth.HostCredentialsWritable) error {
	return nil
}

// ForgetForHost is unimplemented but is required for the auth.CredentialsSource interface.
func (s *CredentialsSource) ForgetForHost(host svchost.Hostname) error {
	return nil
}
