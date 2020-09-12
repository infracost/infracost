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

func parseResourceData(plan, provider, values gjson.Result) map[string]*schema.ResourceData {
	defaultRegion := parseAwsRegion(provider)

	resources := make(map[string]*schema.ResourceData)

	for _, r := range values.Get("resources").Array() {
		t := r.Get("type").String()
		n := r.Get("provider_name").String()
		a := r.Get("address").String()
		v := r.Get("values")

		// Override the region with the region from the arn if exists
		region := defaultRegion
		if v.Get("arn").Exists() {
			region = strings.Split(v.Get("arn").String(), ":")[3]
		}
		v = schema.AddRawValue(v, "region", region)

		resources[a] = schema.NewResourceData(t, n, a, v)
	}

	// Recursively add any resources for child modules
	for _, m := range values.Get("child_modules").Array() {
		resources := parseResourceData(plan, provider, m)
		for address, d := range resources {
			resources[address] = d
		}
	}
	return resources
}

func parseAwsRegion(providerConfig gjson.Result) string {
	// Find region from terraform provider config
	awsRegion := providerConfig.Get("aws.expressions.region.constant_value").String()
	if awsRegion == "" {
		awsRegion = "us-east-1"
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

func getConfigurationJSONForModulePath(conf gjson.Result, names []string) gjson.Result {
	if len(names) == 0 {
		return conf
	}

	// Build up the gjson search key
	p := make([]string, 0, len(names))
	for _, n := range names {
		p = append(p, fmt.Sprintf("module_calls.%s.module", n))
	}

	return conf.Get(strings.Join(p, "."))
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
	p := strings.Split(address, ".")

	if len(p) >= 3 && p[len(p)-3] == "data" {
		return strings.Join(p[len(p)-3:], ".")
	}

	return strings.Join(p[len(p)-2:], ".")
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
	if matches == nil {
		return []string{}
	}

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
