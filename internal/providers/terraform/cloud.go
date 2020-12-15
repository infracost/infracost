package terraform

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/infracost/infracost/internal/config"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var ErrMissingCloudAPIToken = errors.New("No Terraform Cloud API Token is set")
var ErrInvalidCloudAPIToken = errors.New("Invalid Terraform Cloud API Token")

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
		return []byte{}, ErrInvalidCloudAPIToken
	} else if resp.StatusCode != 200 {
		return []byte{}, errors.Errorf("invalid response from Terraform remote: %s", resp.Status)
	}

	return ioutil.ReadAll(resp.Body)
}

func cloudAPIToken(host string) string {
	if config.Config.TerraformCloudAPIToken != "" {
		return config.Config.TerraformCloudAPIToken
	}

	log.Debug("No TERRAFORM_CLOUD_API_TOKEN environment variable set, checking Terraform credential file for matching API token")

	// If the TF_CLI_CONFIG_FILE env variable is set then we shouldn't use the
	// default Terraform credentials file. In the future we may want to support
	// reading the credentials from here as well.
	if os.Getenv("TF_CLI_CONFIG_FILE") != "" {
		log.Debug("TF_CLI_CONFIG_FILE is set, not checking the default credentials file")
		return ""
	}

	credFile := credFile()
	data, err := ioutil.ReadFile(credFile)
	if err != nil {
		log.Debugf("Error reading Terraform credentials file %s: %v", credFile, err)
		return ""
	}

	var parsedCredData struct {
		Credentials map[string]struct {
			Token string
		}
	}
	err = json.Unmarshal(data, &parsedCredData)
	if err != nil {
		log.Debugf("Error parsing Terraform credentials file %s: %v", credFile, err)
		return ""
	}

	if hostCredentials, ok := parsedCredData.Credentials[host]; ok {
		return hostCredentials.Token
	}

	return ""
}

func checkCloudAPITokenSet() bool {
	if config.Config.TerraformCloudAPIToken != "" {
		return true
	}

	// If the TF_CLI_CONFIG_FILE env variable is set then we shouldn't use the
	// default Terraform credentials file. In the future we may want to support
	// reading the credentials from here as well.
	if os.Getenv("TF_CLI_CONFIG_FILE") != "" {
		log.Debug("TF_CLI_CONFIG_FILE is set, not checking the default credentials file")
		return false
	}

	if _, err := os.Stat(credFile()); err == nil {
		return true
	}

	return false
}

func credFile() string {
	var dir string
	if runtime.GOOS == "windows" {
		dir = filepath.Join(os.Getenv("APPDATA"), "terraform.d")
	} else {
		dir, _ = homedir.Expand("~/.terraform.d")
	}
	return path.Join(dir, "credentials.tfrc.json")
}
