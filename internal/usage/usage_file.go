package usage

import (
	"fmt"
	"io/ioutil"

	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v2"
)

const minVersion = "v0.1"
const maxVersion = "v0.1"

type UsageFile struct { // nolint:golint
	Version       string                 `yaml:"version"`
	ResourceUsage map[string]interface{} `yaml:"resource_usage"`
}

func LoadFromFile(usageFile string) (map[string]*schema.UsageData, error) {
	usageData := make(map[string]*schema.UsageData)

	if usageFile == "" {
		return usageData, nil
	}

	log.Debug("Loading usage data from usage file")

	out, err := ioutil.ReadFile(usageFile)
	if err != nil {
		return usageData, errors.Wrapf(err, "Error reading usage file")
	}

	usageData, err = parseYAML(out)
	if err != nil {
		return usageData, errors.Wrapf(err, "Error parsing usage file")
	}

	return usageData, nil
}

func parseYAML(y []byte) (map[string]*schema.UsageData, error) {
	var usageFile UsageFile

	err := yaml.Unmarshal(y, &usageFile)
	if err != nil {
		return map[string]*schema.UsageData{}, errors.Wrap(err, "Error parsing usage YAML")
	}

	if !checkVersion(usageFile.Version) {
		return map[string]*schema.UsageData{}, fmt.Errorf("Invalid usage file version. Supported versions are %s ≤ x ≤ %s", minVersion, maxVersion)
	}

	usageMap := schema.NewUsageMap(usageFile.ResourceUsage)

	return usageMap, nil
}

func checkVersion(v string) bool {
	return semver.Compare(v, minVersion) >= 0 && semver.Compare(v, maxVersion) <= 0
}
