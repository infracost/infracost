package terraform

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"infracost/internal/providers/terraform/aws"
	"infracost/pkg/schema"

	"github.com/tidwall/gjson"
)

func createResource(resourceData *schema.ResourceData) *schema.Resource {
	switch resourceData.Type {
	case "aws_instance":
		return aws.AwsInstance(resourceData)
	case "aws_nat_gateway":
		return aws.AwsNatGateway(resourceData)
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

	for _, resourceData := range resourceDataMap {
		resource := createResource(resourceData)
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
		address := terraformResource.Get("address").String()
		resourceType := terraformResource.Get("type").String()
		rawValues := terraformResource.Get("values")

		// Override the region with the region from the arn if it
		awsRegion := defaultAwsRegion
		if rawValues.Get("arn").Exists() {
			awsRegion = strings.Split(rawValues.Get("arn").String(), ":")[3]
		}
		rawValues = addRawValue(rawValues, "region", awsRegion)

		resourceDataMap[address] = schema.NewResourceData(resourceType, address, rawValues)
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

func addRawValue(rawValues gjson.Result, key string, value interface{}) gjson.Result {
	var unmarshalledJSON map[string]interface{}
	json.Unmarshal([]byte(rawValues.Raw), &unmarshalledJSON)
	unmarshalledJSON[key] = value
	marshalledJSON, _ := json.Marshal(unmarshalledJSON)
	return gjson.ParseBytes(marshalledJSON)
}

func parseReferences(resourceDataMap map[string]*schema.ResourceData, configurationJSON gjson.Result) {
	for address, resourceData := range resourceDataMap {
		resourceConfigJSON := getConfigurationJSONForResourceAddress(configurationJSON, address)

		var refAddressesMap = make(map[string][]string)
		for attribute, attributeJSON := range resourceConfigJSON.Get("expressions").Map() {
			getReferenceAddresses(resourceData, attribute, attributeJSON, &refAddressesMap)
		}

		for attribute, refAddresses := range refAddressesMap {
			for _, refAddress := range refAddresses {
				fullRefAddress := fmt.Sprintf("%s.%s", addressModulePart(address), refAddress)
				if refResourceData, ok := resourceDataMap[fullRefAddress]; ok {
					resourceData.AddReference(attribute, refResourceData)
				}
			}
		}
	}
}

func getReferenceAddresses(resourceData *schema.ResourceData, attribute string, attributeJSON gjson.Result, refAddressesMap *map[string][]string) {
	if attributeJSON.Get("references").Exists() {
		for _, ref := range attributeJSON.Get("references").Array() {
			if _, ok := (*refAddressesMap)[attribute]; !ok {
				(*refAddressesMap)[attribute] = make([]string, 0, 1)
			}
			(*refAddressesMap)[attribute] = append((*refAddressesMap)[attribute], ref.String())
		}
	} else if attributeJSON.IsArray() {
		for i, attributeJSONItem := range attributeJSON.Array() {
			getReferenceAddresses(resourceData, fmt.Sprintf("%s.%d", attribute, i), attributeJSONItem, refAddressesMap)
		}
	} else if attributeJSON.Type.String() == "JSON" {
		attributeJSON.ForEach(func(childAttribute gjson.Result, childAttributeJSON gjson.Result) bool {
			getReferenceAddresses(resourceData, fmt.Sprintf("%s.%s", attribute, childAttribute), childAttributeJSON, refAddressesMap)
			return true
		})
	}
}

func getConfigurationJSONForResourceAddress(configurationJSON gjson.Result, address string) gjson.Result {
	moduleNames := addressModuleNames(address)
	moduleConfigJSON := getConfigurationJSONForModulePath(configurationJSON, moduleNames)
	resourceKey := fmt.Sprintf(`resources.#(address="%s")`, stripAddressArray(addressResourcePart(address)))
	return moduleConfigJSON.Get(resourceKey)
}

func getConfigurationJSONForModulePath(configurationJSON gjson.Result, moduleNames []string) gjson.Result {
	// Build up the gjson search key
	moduleKeyParts := make([]string, 0, len(moduleNames))
	for _, moduleName := range moduleNames {
		moduleKeyParts = append(moduleKeyParts, fmt.Sprintf("module_calls.%s.module", moduleName))
	}
	moduleKey := strings.Join(moduleKeyParts, ".")

	return configurationJSON.Get(moduleKey)
}

func addressResourcePart(address string) string {
	addressParts := strings.Split(address, ".")
	resourceParts := addressParts[len(addressParts)-2:]
	return strings.Join(resourceParts, ".")
}

func addressModulePart(address string) string {
	addressParts := strings.Split(address, ".")
	moduleParts := addressParts[:len(addressParts)-2]
	return strings.Join(moduleParts, ".")
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

func stripAddressArray(address string) string {
	addressParts := strings.Split(address, ".")
	resourceParts := addressParts[len(addressParts)-2:]

	r := regexp.MustCompile(`([^\[]+)`)
	match := r.FindStringSubmatch(strings.Join(resourceParts, "."))
	return match[1]
}
