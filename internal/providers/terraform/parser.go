package terraform

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/infracost/infracost/pkg/schema"

	"github.com/tidwall/gjson"
)

// These show differently in the plan JSON for Terraform 0.12 and 0.13
var infracostProviderNames = []string{"infracost", "infracost.io/infracost/infracost"}

func createResource(resourceData *schema.ResourceData, usageData *schema.ResourceData) *schema.Resource {
	resourceRegistry := getResourceRegistry()
	if rFunc, ok := (*resourceRegistry)[resourceData.Type]; ok {
		return rFunc(resourceData, usageData)
	}
	return nil
}

func parsePlanJSON(j []byte) []*schema.Resource {
	planJSON := gjson.ParseBytes(j)
	providerConfig := planJSON.Get("configuration.provider_config")
	plannedValuesJSON := planJSON.Get("planned_values.root_module")
	configurationJSON := planJSON.Get("configuration.root_module")

	resources := make([]*schema.Resource, 0)

	resourceDataMap := parseResourceData(planJSON, providerConfig, plannedValuesJSON)
	parseReferences(resourceDataMap, configurationJSON)
	usageResourceDataMap := buildUsageResourceDataMap(resourceDataMap)
	resourceDataMap = stripInfracostResources(resourceDataMap)

	for _, resourceData := range resourceDataMap {
		usageResourceData := usageResourceDataMap[resourceData.Address]
		resource := createResource(resourceData, usageResourceData)
		if resource != nil {
			resources = append(resources, resource)
		}
	}

	return resources
}

func parseResourceData(planJSON gjson.Result, providerConfig gjson.Result, plannedValuesJSON gjson.Result) map[string]*schema.ResourceData {
	defaultAwsRegion := parseAwsRegion(providerConfig)

	resourceDataMap := make(map[string]*schema.ResourceData)

	for _, terraformResource := range plannedValuesJSON.Get("resources").Array() {
		resourceType := terraformResource.Get("type").String()
		providerName := terraformResource.Get("provider_name").String()
		address := terraformResource.Get("address").String()
		rawValues := terraformResource.Get("values")

		// Override the region with the region from the arn if it
		awsRegion := defaultAwsRegion
		if rawValues.Get("arn").Exists() {
			awsRegion = strings.Split(rawValues.Get("arn").String(), ":")[3]
		}
		rawValues = schema.AddRawValue(rawValues, "region", awsRegion)

		resourceDataMap[address] = schema.NewResourceData(resourceType, providerName, address, rawValues)
	}

	// Recursively add any resources for child modules
	for _, modulePlannedValueJSON := range plannedValuesJSON.Get("child_modules").Array() {
		moduleResourceDataMap := parseResourceData(planJSON, providerConfig, modulePlannedValueJSON)
		for address, resourceData := range moduleResourceDataMap {
			resourceDataMap[address] = resourceData
		}
	}
	return resourceDataMap
}

func parseAwsRegion(providerConfig gjson.Result) string {
	awsRegion := "us-east-1" // Use as fallback

	// Find region from terraform provider config
	awsRegionConfig := providerConfig.Get("aws.expressions.region.constant_value").String()
	if awsRegionConfig != "" {
		awsRegion = awsRegionConfig
	}

	return awsRegion
}

func buildUsageResourceDataMap(resourceDataMap map[string]*schema.ResourceData) map[string]*schema.ResourceData {
	usageResourceDataMap := make(map[string]*schema.ResourceData)
	for _, resourceData := range resourceDataMap {
		if isInfracostResource(resourceData) {
			for _, refResourceData := range resourceData.References("resources") {
				usageResourceDataMap[refResourceData.Address] = resourceData
			}
		}
	}
	return usageResourceDataMap
}

func stripInfracostResources(resourceDataMap map[string]*schema.ResourceData) map[string]*schema.ResourceData {
	newResourceDataMap := make(map[string]*schema.ResourceData)
	for address, resourceData := range resourceDataMap {
		if !isInfracostResource(resourceData) {
			newResourceDataMap[address] = resourceData
		}
	}
	return newResourceDataMap
}

func parseReferences(resourceDataMap map[string]*schema.ResourceData, configurationJSON gjson.Result) {
	for address, resourceData := range resourceDataMap {
		resourceConfigJSON := getConfigurationJSONForResourceAddress(configurationJSON, address)

		var referencesMap = make(map[string][]string)
		for attribute, attributeJSON := range resourceConfigJSON.Get("expressions").Map() {
			getReferences(resourceData, attribute, attributeJSON, &referencesMap)
		}

		resourceCountIndex := addressCountIndex(address)

		for attribute, references := range referencesMap {
			referenceHasCount := containsString(references, "count.index")
			for _, reference := range references {
				if reference == "count.index" {
					continue
				}
				arrayPart := ""
				if referenceHasCount {
					arrayPart = fmt.Sprintf("[%d]", resourceCountIndex)
				}
				fullRefAddress := fmt.Sprintf("%s%s%s", addressModulePart(address), reference, arrayPart)
				if refResourceData, ok := resourceDataMap[fullRefAddress]; ok {
					resourceData.AddReference(attribute, refResourceData)
				}
			}
		}
	}
}

func getReferences(resourceData *schema.ResourceData, attribute string, attributeJSON gjson.Result, referencesMap *map[string][]string) {
	if attributeJSON.Get("references").Exists() {
		for _, ref := range attributeJSON.Get("references").Array() {
			if _, ok := (*referencesMap)[attribute]; !ok {
				(*referencesMap)[attribute] = make([]string, 0, 1)
			}
			(*referencesMap)[attribute] = append((*referencesMap)[attribute], ref.String())
		}
	} else if attributeJSON.IsArray() {
		for i, attributeJSONItem := range attributeJSON.Array() {
			getReferences(resourceData, fmt.Sprintf("%s.%d", attribute, i), attributeJSONItem, referencesMap)
		}
	} else if attributeJSON.Type.String() == "JSON" {
		attributeJSON.ForEach(func(childAttribute gjson.Result, childAttributeJSON gjson.Result) bool {
			getReferences(resourceData, fmt.Sprintf("%s.%s", attribute, childAttribute), childAttributeJSON, referencesMap)
			return true
		})
	}
}

func getConfigurationJSONForResourceAddress(configurationJSON gjson.Result, address string) gjson.Result {
	moduleNames := addressModuleNames(address)
	moduleConfigJSON := getConfigurationJSONForModulePath(configurationJSON, moduleNames)
	resourceKey := fmt.Sprintf(`resources.#(address="%s")`, removeAddressArrayPart(addressResourcePart(address)))
	return moduleConfigJSON.Get(resourceKey)
}

func getConfigurationJSONForModulePath(configurationJSON gjson.Result, moduleNames []string) gjson.Result {
	// Build up the gjson search key
	moduleKeyParts := make([]string, 0, len(moduleNames))
	for _, moduleName := range moduleNames {
		moduleKeyParts = append(moduleKeyParts, fmt.Sprintf("module_calls.%s.module", moduleName))
	}

	if len(moduleKeyParts) == 0 {
		return configurationJSON
	} else {
		moduleKey := strings.Join(moduleKeyParts, ".")
		return configurationJSON.Get(moduleKey)
	}
}

func isInfracostResource(resourceData *schema.ResourceData) bool {
	for _, providerName := range infracostProviderNames {
		if resourceData.ProviderName == providerName {
			return true
		}
	}
	return false
}

func addressResourcePart(address string) string {
	addressParts := strings.Split(address, ".")
	var resourceParts []string
	if len(addressParts) >= 3 && addressParts[len(addressParts)-3] == "data" {
		resourceParts = addressParts[len(addressParts)-3:]
	} else {
		resourceParts = addressParts[len(addressParts)-2:]
	}
	return strings.Join(resourceParts, ".")
}

func addressModulePart(address string) string {
	addressParts := strings.Split(address, ".")
	var moduleParts []string
	if len(addressParts) >= 3 && addressParts[len(addressParts)-3] == "data" {
		moduleParts = addressParts[:len(addressParts)-3]
	} else {
		moduleParts = addressParts[:len(addressParts)-2]
	}
	if len(moduleParts) == 0 {
		return ""
	} else {
		return fmt.Sprintf("%s.", strings.Join(moduleParts, "."))
	}
}

func addressModuleNames(address string) []string {
	r := regexp.MustCompile(`module\.([^\[]*)`)
	matches := r.FindAllStringSubmatch(addressModulePart(address), -1)

	moduleNames := make([]string, 0, len(matches))
	for _, match := range matches {
		moduleNames = append(moduleNames, match[1])
	}

	return moduleNames
}

func addressCountIndex(address string) int {
	r := regexp.MustCompile(`\[(\d+)\]`)
	match := r.FindStringSubmatch(address)
	if len(match) > 0 {
		i, _ := strconv.Atoi(match[1])
		return i
	}
	return -1
}

func removeAddressArrayPart(address string) string {
	r := regexp.MustCompile(`([^\[]+)`)
	match := r.FindStringSubmatch(addressResourcePart(address))
	return match[1]
}

func containsString(arr []string, s string) bool {
	for _, item := range arr {
		if item == s {
			return true
		}
	}
	return false
}
