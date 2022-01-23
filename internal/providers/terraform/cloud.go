package terraform

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hclparse"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var ErrMissingCloudToken = errors.New("no Terraform Cloud token is set")
var ErrInvalidCloudToken = errors.New("invalid Terraform Cloud Token")

func cloudAPI(host string, path string, token string) ([]byte, error) {
	client := &http.Client{}

	url := fmt.Sprintf("https://%s%s", host, path)
	log.Debugf("Calling Terraform Cloud API: %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []byte{}, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return []byte{}, ErrInvalidCloudToken
	} else if resp.StatusCode != 200 {
		return []byte{}, errors.Errorf("invalid response from Terraform remote: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func findCloudToken(host string) string {
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

func checkCloudConfigSet() bool {
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
