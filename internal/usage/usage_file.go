package usage

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/infracost/infracost"
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

func SyncUsageData(project *schema.Project, existingUsageData map[string]*schema.UsageData, usageFilePath string) error {
	if usageFilePath == "" {
		return nil
	}
	usageSchema, err := loadUsageSchema()
	if err != nil {
		return err
	}
	syncedResourcesUsage := syncResourcesUsage(project.Resources, usageSchema, existingUsageData)
	// yaml.MapSlice is used to maintain the order of keys, so re-running
	// the code won't change the output.
	syncedUsageData := yaml.MapSlice{
		{Key: "version", Value: 0.1},
		{Key: "resource_usage", Value: syncedResourcesUsage},
	}
	d, err := yaml.Marshal(syncedUsageData)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(usageFilePath, d, 0600)
	if err != nil {
		return err
	}
	return nil
}

func syncResourcesUsage(resources []*schema.Resource, usageSchema map[string][]string, existingUsageData map[string]*schema.UsageData) yaml.MapSlice {
	syncedResourceUsage := make(map[string]interface{})
	for _, resource := range resources {
		resourceName := resource.Name
		resourceTypeName := strings.Split(resourceName, ".")[0]
		resourceUSchema, ok := usageSchema[resourceTypeName]
		if !ok {
			continue
		}
		resourceUsage := make(map[string]int64)
		for _, usageKey := range resourceUSchema {
			var usageValue int64 = 0
			if existingUsage, ok := existingUsageData[resourceName]; ok {
				usageValue = existingUsage.Get(usageKey).Int()
			}
			resourceUsage[usageKey] = usageValue
		}
		syncedResourceUsage[resourceName] = unFlattenHelper(resourceUsage)
	}
	// yaml.MapSlice is used to maintain the order of keys, so re-running
	// the code won't change the output.
	result := mapToSortedMapSlice(syncedResourceUsage)
	return result
}

func loadUsageSchema() (map[string][]string, error) {
	usageSchema := make(map[string][]string)
	usageData, err := loadReferenceFile()
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

func unFlattenHelper(input map[string]int64) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range input {
		rootMap := &result
		splittedKey := strings.Split(k, ".")
		for it := 0; it < len(splittedKey)-1; it++ {
			key := splittedKey[it]
			if _, ok := (*rootMap)[key]; !ok {
				(*rootMap)[key] = make(map[string]interface{})
			}
			casted := (*rootMap)[key].(map[string]interface{})
			rootMap = &casted
		}
		key := splittedKey[len(splittedKey)-1]
		(*rootMap)[key] = v
	}
	return result
}

func mapToSortedMapSlice(input map[string]interface{}) yaml.MapSlice {
	result := make(yaml.MapSlice, 0)
	// sort keys of the input to maintain same output for different runs.
	keys := make([]string, 0)
	for k := range input {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	// Iterate over sorted keys
	for _, k := range keys {
		v := input[k]
		if casted, ok := v.(map[string]interface{}); ok {
			result = append(result, yaml.MapItem{Key: k, Value: mapToSortedMapSlice(casted)})
		} else {
			result = append(result, yaml.MapItem{Key: k, Value: v})
		}
	}
	return result
}

func loadReferenceFile() (map[string]*schema.UsageData, error) {
	referenceUsageFileContents := infracost.GetReferenceUsageFileContents()
	usageData, err := parseYAML(*referenceUsageFileContents)
	if err != nil {
		return usageData, errors.Wrapf(err, "Error parsing usage file")
	}
	return usageData, nil
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
