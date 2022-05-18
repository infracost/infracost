package modules

import (
	"net/http"

	svchost "github.com/hashicorp/terraform-svchost"
	"github.com/hashicorp/terraform-svchost/auth"
)

type FindTokenForHost func(host string) (string, error)

type HostCredentials struct {
	token string
}

func NewHostCredentials(token string) *HostCredentials {
	return &HostCredentials{
		token: token,
	}
}

func (c *HostCredentials) Token() string {
	return c.token
}

func (c *HostCredentials) PrepareRequest(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.Token())
}

type CredentialsSource struct {
	findTokenForHost FindTokenForHost
}

func NewCredentialsSource(findTokenForHost FindTokenForHost) *CredentialsSource {
	return &CredentialsSource{
		findTokenForHost: findTokenForHost,
	}
}

func (s *CredentialsSource) ForHost(host svchost.Hostname) (auth.HostCredentials, error) {
	token, err := s.findTokenForHost(host.ForDisplay())
	if err != nil {
		return nil, err
	}

	return NewHostCredentials(token), nil
}

func (s *CredentialsSource) StoreForHost(host svchost.Hostname, credentials auth.HostCredentialsWritable) error {
	return nil
}

func (s *CredentialsSource) ForgetForHost(host svchost.Hostname) error {
	return nil
}
