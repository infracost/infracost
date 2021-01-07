package usage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
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
	usageMap := make(map[string]*schema.UsageData)

	var usageFile UsageFile

	err := yaml.Unmarshal(y, &usageFile)
	if err != nil {
		return usageMap, errors.Wrap(err, "Error parsing usage YAML")
	}

	if !checkVersion(usageFile.Version) {
		return usageMap, fmt.Errorf("Invalid usage file version. Supported versions are %s ≤ x ≤ %s", minVersion, maxVersion)
	}

	for addr, v := range usageFile.ResourceUsage {
		usageMap[addr] = schema.NewUsageData(
			addr,
			ParseAttributes(v),
		)
	}

	return usageMap, nil
}

func ParseAttributes(i interface{}) map[string]gjson.Result {
	a := make(map[string]gjson.Result)
	for k, v := range flatten(i) {
		j, _ := json.Marshal(v)
		a[k] = gjson.ParseBytes(j)
	}

	return a
}

func checkVersion(v string) bool {
	return semver.Compare(v, minVersion) >= 0 && semver.Compare(v, maxVersion) <= 0
}

func flatten(i interface{}) map[string]interface{} {
	keys := make([]string, 0)
	result := make(map[string]interface{})
	flattenHelper(i, keys, result)
	return result
}

func flattenHelper(i interface{}, keys []string, result map[string]interface{}) {
	switch v := i.(type) {
	case map[interface{}]interface{}:
		for k, v := range i.(map[interface{}]interface{}) {
			flattenHelper(v, append(keys, fmt.Sprintf("%v", k)), result)
		}
	default:
		k := strings.Join(keys, ".")
		result[k] = v
	}
}
