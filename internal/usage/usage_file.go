package usage

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v2"
)

const minUsageFileVersion = "0.1"
const maxUsageFileVersion = "0.1"

type UsageFile struct { // nolint:golint
	Version       string                 `yaml:"version"`
	ResourceUsage map[string]interface{} `yaml:"resource_usage"`
}

func SyncUsageData(project *schema.Project, usageData map[string]*schema.UsageData) error {
	usageSchema, err := loadUsageSchema()
	if err != nil {
		return err
	}
	syncedUsageData := make(map[string]map[string]int64)
	for _, resource := range project.Resources {
		resourceName := resource.Name
		resourceTypeName := strings.Split(resourceName, ".")[0]

		// TODO: Move to another function.
		resourceUSchema, ok := usageSchema[resourceTypeName]
		if !ok {
			continue
		}
		syncedUsageData[resourceName] = make(map[string]int64)
		for _, usageKey := range resourceUSchema {
			syncedUsageData[resourceName][usageKey] = 0
			if existingUsage, ok := usageData[resourceName]; ok {
				syncedUsageData[resourceName][usageKey] = existingUsage.Get(usageKey).Int()
			}
		}
	}
	d, err := yaml.Marshal(syncedUsageData)
	if err != nil {
		return err
	}
	fmt.Println(string(d))
	return nil
}

func loadUsageSchema() (map[string][]string, error) {
	usageSchema := make(map[string][]string)
	// TODO: How to ship the schema in the production binary?
	usageData, err := LoadFromFile("infracost-usage-example.yml")
	if err != nil {
		return usageSchema, err
	}
	for _, resUsageData := range usageData {
		resourceTypeName := strings.Split(resUsageData.Address, ".")[0]
		usageSchema[resourceTypeName] = make([]string, 0)
		for usageKeyName := range resUsageData.Attributes {
			usageSchema[resourceTypeName] = append(usageSchema[resourceTypeName], usageKeyName)
		}
	}
	return usageSchema, nil
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
		return map[string]*schema.UsageData{}, fmt.Errorf("Invalid usage file version. Supported versions are %s ≤ x ≤ %s", minUsageFileVersion, maxUsageFileVersion)
	}

	usageMap := schema.NewUsageMap(usageFile.ResourceUsage)

	return usageMap, nil
}

func checkVersion(v string) bool {
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	return semver.Compare(v, "v"+minUsageFileVersion) >= 0 && semver.Compare(v, "v"+maxUsageFileVersion) <= 0
}
