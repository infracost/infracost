package usage

import (
	"fmt"
	"io/ioutil"
	"os"
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

type SchemaItem struct {
	Key          string
	ValueType    schema.UsageVariableType
	DefaultValue interface{}
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

func syncResourcesUsage(resources []*schema.Resource, usageSchema map[string][]*SchemaItem, existingUsageData map[string]*schema.UsageData) yaml.MapSlice {
	syncedResourceUsage := make(map[string]interface{})
	for _, resource := range resources {
		resourceName := resource.Name
		resourceUSchema := resource.UsageSchema
		if resource.UsageSchema == nil {
			// There is no explicitly defined UsageSchema for this resource.  Use the old way and create one from
			// infracost-usage-example.yml.
			resourceTypeNames := strings.Split(resourceName, ".")
			if len(resourceTypeNames) < 2 {
				// It's a resource with no name
				continue
			}
			// This handles module names appearing in the resourceName too
			resourceTypeName := resourceTypeNames[len(resourceTypeNames)-2]
			schemaItems, ok := usageSchema[resourceTypeName]
			if !ok {
				continue
			}

			resourceUSchema = make([]*schema.UsageSchemaItem, 0, len(schemaItems))
			for _, s := range schemaItems {
				resourceUSchema = append(resourceUSchema, &schema.UsageSchemaItem{
					Key:          s.Key,
					DefaultValue: s.DefaultValue,
					ValueType:    s.ValueType,
					ShouldSync:   true,
				})
			}
		}

		resourceUsage := make(map[string]interface{})
		for _, usageSchemaItem := range resourceUSchema {
			usageKey := usageSchemaItem.Key
			usageValueType := usageSchemaItem.ValueType
			var existingUsageValue interface{}
			if existingUsage, ok := existingUsageData[resourceName]; ok {
				switch usageValueType {
				case schema.Float64:
					if v := existingUsage.GetFloat(usageKey); v != nil {
						existingUsageValue = *v
					}
				case schema.Int64:
					if v := existingUsage.GetInt(usageKey); v != nil {
						existingUsageValue = *v
					}
				case schema.String:
					if v := existingUsage.GetString(usageKey); v != nil {
						existingUsageValue = *v
					}
				case schema.StringArray:
					if v := existingUsage.GetStringArray(usageKey); v != nil {
						existingUsageValue = *v
					}
				}
			}
			if existingUsageValue != nil {
				resourceUsage[usageKey] = existingUsageValue
			} else if usageSchemaItem.ShouldSync {
				resourceUsage[usageKey] = usageSchemaItem.DefaultValue
			}
		}
		syncedResourceUsage[resourceName] = unFlattenHelper(resourceUsage)
	}
	// yaml.MapSlice is used to maintain the order of keys, so re-running
	// the code won't change the output.
	result := mapToSortedMapSlice(syncedResourceUsage)
	return result
}

func loadUsageSchema() (map[string][]*SchemaItem, error) {
	usageSchema := make(map[string][]*SchemaItem)
	usageData, err := loadReferenceFile()
	if err != nil {
		return usageSchema, err
	}
	for _, resUsageData := range usageData {
		resourceTypeName := strings.Split(resUsageData.Address, ".")[0]
		usageSchema[resourceTypeName] = make([]*SchemaItem, 0)
		for usageKeyName, usageRawResult := range resUsageData.Attributes {
			var defaultValue interface{}
			usageValueType := schema.Int64
			defaultValue = 0
			usageRawValue := usageRawResult.Value()
			if _, ok := usageRawValue.(string); ok {
				usageValueType = schema.String
				defaultValue = usageRawResult.String()
			}
			usageSchema[resourceTypeName] = append(usageSchema[resourceTypeName], &SchemaItem{
				Key:          usageKeyName,
				ValueType:    usageValueType,
				DefaultValue: defaultValue,
			})
		}
	}
	return usageSchema, nil
}

func unFlattenHelper(input map[string]interface{}) map[string]interface{} {
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

func LoadFromFile(usageFilePath string, createIfNotExisting bool) (map[string]*schema.UsageData, error) {
	usageData := make(map[string]*schema.UsageData)

	if usageFilePath == "" {
		return usageData, nil
	}

	if createIfNotExisting {
		if _, err := os.Stat(usageFilePath); os.IsNotExist(err) {
			log.Debug("Specified usage file does not exist. It will be created")
			fileContent := yaml.MapSlice{
				{Key: "version", Value: "0.1"},
				{Key: "resource_usage", Value: make(map[string]interface{})},
			}
			d, err := yaml.Marshal(fileContent)
			if err != nil {
				return usageData, errors.Wrapf(err, "Error creating usage file")
			}
			err = ioutil.WriteFile(usageFilePath, d, 0600)
			if err != nil {
				return usageData, errors.Wrapf(err, "Error creating usage file")
			}
		}
	}

	log.Debug("Loading usage data from usage file")

	out, err := ioutil.ReadFile(usageFilePath)
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
