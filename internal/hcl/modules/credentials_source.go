package modules

import (
	"net/http"

	svchost "github.com/hashicorp/terraform-svchost"
	"github.com/hashicorp/terraform-svchost/auth"

	"github.com/infracost/infracost/internal/credentials"
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

// CredentialsSource is an object that may be able to provide credentials
// for a given module source.
type CredentialsSource struct {
	BaseCredentialSet BaseCredentialSet
	FetchToken        FetchTokenFunc
}

// FetchTokenFunc defines a function that returns a token for a given key.
// This can be an environment key, a header key, whatever the CredentialsSource requires.
type FetchTokenFunc func(key string) string

// BaseCredentialSet are the underlying credentials that CredentialsSource will use if no other credentials can be
// found for a given host.
type BaseCredentialSet struct {
	Token string
	Host  string
}

// NewTerraformCredentialsSource returns a CredentialsSource attempting to set the BaseCredentialSet as the base.
// If creds doesn't contain static details, NewTerraformCredentialsSource will attempt to fill the credential set from
// the environment, returning an error if it cannot.
func NewTerraformCredentialsSource(creds BaseCredentialSet) (*CredentialsSource, error) {
	if creds.Host == "" {
		creds.Host = "app.terraform.io"
	}

	f := credentials.FindTerraformCloudToken
	c := &CredentialsSource{FetchToken: f}

	if creds.Token == "" && !credentials.CheckCloudConfigSet() {
		return c, credentials.ErrMissingCloudToken
	}

	if creds.Token == "" {
		creds.Token = f(creds.Host)
	}

	if creds.Token == "" {
		return c, credentials.ErrMissingCloudToken
	}

	return &CredentialsSource{
		BaseCredentialSet: creds,
		FetchToken:        f,
	}, nil
}

// ForHost returns a non-nil HostCredentials if the source has credentials
// available for the host, and a nil HostCredentials if it does not.
func (s *CredentialsSource) ForHost(host svchost.Hostname) (auth.HostCredentials, error) {
	display := host.ForDisplay()
	if s.BaseCredentialSet.Host == display {
		return HostCredentials{token: s.BaseCredentialSet.Token}, nil
	}

	return HostCredentials{token: s.FetchToken(display)}, nil
}

// StoreForHost is unimplemented but is required for the auth.CredentialsSource interface.
func (s *CredentialsSource) StoreForHost(host svchost.Hostname, credentials auth.HostCredentialsWritable) error {
	return nil
}

// ForgetForHost is unimplemented but is required for the auth.CredentialsSource interface.
func (s *CredentialsSource) ForgetForHost(host svchost.Hostname) error {
	return nil
}
