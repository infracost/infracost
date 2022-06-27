package modules

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	svchost "github.com/hashicorp/terraform-svchost"
	"github.com/hashicorp/terraform-svchost/auth"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
)

var (
	ErrMissingCloudToken = errors.New("no Terraform Cloud token is set")
	ErrInvalidCloudToken = errors.New("invalid Terraform Cloud Token")
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
}

// BaseCredentialSet are the underlying credentials that CredentialsSource will use if no other credentials can be
// found for a given host.
type BaseCredentialSet struct {
	TerraformCloudToken string
	TerraformCloudHost  string
}

// NewCredentialsSource returns a CredentialsSource attempting to set the BaseCredentialSet as the base.
// If creds doesn't contain static details, NewCredentialsSource will attempt to fill the credential set from
// the environment, returning an error if it cannot.
func NewCredentialsSource(creds BaseCredentialSet) (*CredentialsSource, error) {
	if creds.TerraformCloudHost == "" {
		creds.TerraformCloudHost = "app.terraform.io"
	}

	if creds.TerraformCloudToken == "" && !CheckCloudConfigSet() {
		return &CredentialsSource{}, ErrMissingCloudToken
	}

	if creds.TerraformCloudToken == "" {
		creds.TerraformCloudToken = FindTerraformCloudToken(creds.TerraformCloudHost)
	}

	if creds.TerraformCloudToken == "" {
		return &CredentialsSource{}, ErrMissingCloudToken
	}

	return &CredentialsSource{
		BaseCredentialSet: creds,
	}, nil
}

// ForHost returns a non-nil HostCredentials if the source has credentials
// available for the host, and a nil HostCredentials if it does not.
func (s *CredentialsSource) ForHost(host svchost.Hostname) (auth.HostCredentials, error) {
	display := host.ForDisplay()
	if s.BaseCredentialSet.TerraformCloudHost == display {
		return HostCredentials{token: s.BaseCredentialSet.TerraformCloudToken}, nil
	}

	return HostCredentials{token: FindTerraformCloudToken(display)}, nil
}

// StoreForHost is unimplemented but is required for the auth.CredentialsSource interface.
func (s *CredentialsSource) StoreForHost(host svchost.Hostname, credentials auth.HostCredentialsWritable) error {
	return nil
}

// ForgetForHost is unimplemented but is required for the auth.CredentialsSource interface.
func (s *CredentialsSource) ForgetForHost(host svchost.Hostname) error {
	return nil
}

// FindTerraformCloudToken returns a TFC Bearer token for the given host.
func FindTerraformCloudToken(host string) string {
	if os.Getenv("TF_CLI_CONFIG_FILE") != "" {
		log.Debugf("TF_CLI_CONFIG_FILE is set, checking %s for Terraform Cloud credentials", os.Getenv("TF_CLI_CONFIG_FILE"))
		token, err := credFromHCL(os.Getenv("TF_CLI_CONFIG_FILE"), host)
		if err != nil {
			log.Debugf("Error reading Terraform config file %s: %v", os.Getenv("TF_CLI_CONFIG_FILE"), err)
		}
		if token != "" {
			return token
		}
	}

	credFile := defaultCredFile()
	if _, err := os.Stat(credFile); err == nil {
		log.Debugf("Checking %s for Terraform Cloud credentials", credFile)
		token, err := credFromJSON(credFile, host)
		if err != nil {
			log.Debugf("Error reading Terraform credentials file %s: %v", credFile, err)
		}
		if token != "" {
			return token
		}
	}

	confFile := defaultConfFile()
	if _, err := os.Stat(confFile); err == nil {
		log.Debugf("Checking %s for Terraform Cloud credentials", confFile)
		token, err := credFromHCL(confFile, host)
		if err != nil {
			log.Debugf("Error reading Terraform config file %s: %v", confFile, err)
		}
		if token != "" {
			return token
		}
	}

	return ""
}

func credFromHCL(filename string, host string) (string, error) {
	parser := hclparse.NewParser()
	f, parseDiags := parser.ParseHCLFile(filename)
	if parseDiags.HasErrors() {
		return "", parseDiags
	}

	var conf struct {
		Credentials []struct {
			Name  string `hcl:"name,label"`
			Token string `hcl:"token"`
		} `hcl:"credentials,block"`
	}

	decodeDiags := gohcl.DecodeBody(f.Body, nil, &conf)
	if decodeDiags.HasErrors() {
		return "", parseDiags
	}

	for _, c := range conf.Credentials {
		if c.Name == host {
			return c.Token, nil
		}
	}

	return "", nil
}

func credFromJSON(filename, host string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	var conf struct {
		Credentials map[string]struct {
			Token string `json:"token"`
		} `json:"credentials"`
	}
	err = json.Unmarshal(data, &conf)
	if err != nil {
		return "", err
	}

	if hostCred, ok := conf.Credentials[host]; ok {
		return hostCred.Token, nil
	}

	return "", nil
}

// CheckCloudConfigSet returns if there is valid configuration for Terraform Cloud available on the system.
func CheckCloudConfigSet() bool {
	if os.Getenv("TF_CLI_CONFIG_FILE") != "" {
		return true
	}

	if _, err := os.Stat(defaultConfFile()); err == nil {
		return true
	}

	if _, err := os.Stat(defaultCredFile()); err == nil {
		return true
	}

	return false
}

func defaultConfFile() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "terraform.rc")
	}

	p, _ := homedir.Expand("~/.terraformrc")
	return p
}

func defaultCredFile() string {
	var dir string
	if runtime.GOOS == "windows" {
		dir = filepath.Join(os.Getenv("APPDATA"), "terraform.d")
	} else {
		dir, _ = homedir.Expand("~/.terraform.d")
	}
	return path.Join(dir, "credentials.tfrc.json")
}
