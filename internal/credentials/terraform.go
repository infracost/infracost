package credentials

import (
	"encoding/json"
	"errors"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/mitchellh/go-homedir"

	"github.com/infracost/infracost/internal/logging"
)

var (
	ErrMissingCloudToken = errors.New("no Terraform Cloud token is set")
	ErrInvalidCloudToken = errors.New("invalid Terraform Cloud Token")
)

// FindTerraformCloudToken returns a TFC Bearer token for the given host.
func FindTerraformCloudToken(host string) string {
	if os.Getenv("TF_CLI_CONFIG_FILE") != "" {
		logging.Logger.Debug().Msgf("TF_CLI_CONFIG_FILE is set, checking %s for Terraform Cloud credentials", os.Getenv("TF_CLI_CONFIG_FILE"))
		token, err := credFromHCL(os.Getenv("TF_CLI_CONFIG_FILE"), host)
		if err != nil {
			logging.Logger.Debug().Msgf("Error reading Terraform config file %s: %v", os.Getenv("TF_CLI_CONFIG_FILE"), err)
		}
		if token != "" {
			return token
		}
	}

	credFile := defaultCredFile()
	if _, err := os.Stat(credFile); err == nil {
		logging.Logger.Debug().Msgf("Checking %s for Terraform Cloud credentials", credFile)
		token, err := credFromJSON(credFile, host)
		if err != nil {
			logging.Logger.Debug().Msgf("Error reading Terraform credentials file %s: %v", credFile, err)
		}
		if token != "" {
			return token
		}
	}

	confFile := defaultConfFile()
	if _, err := os.Stat(confFile); err == nil {
		logging.Logger.Debug().Msgf("Checking %s for Terraform Cloud credentials", confFile)
		token, err := credFromHCL(confFile, host)
		if err != nil {
			logging.Logger.Debug().Msgf("Error reading Terraform config file %s: %v", confFile, err)
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
