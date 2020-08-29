package terraform

import (
	"fmt"
	"strings"

	"infracost/internal/terraform/aws"
	"infracost/pkg/resource"

	"github.com/tidwall/gjson"
)

func createResource(resourceType string, address string, rawValues map[string]interface{}, providerConfig gjson.Result) resource.Resource {
	awsRegion := "us-east-1" // Use as fallback

	// Find region from terraform provider config
	awsRegionConfig := providerConfig.Get("aws.expressions.region.constant_value").String()
	if awsRegionConfig != "" {
		awsRegion = awsRegionConfig
	}

	// Override the region with the region from the arn if it
	arn := rawValues["arn"]
	if arn != nil {
		awsRegion = strings.Split(arn.(string), ":")[3]
	}

	switch resourceType {
	// case "aws_instance":
	// 	return aws.NewEc2Instance(address, awsRegion, rawValues)
	// case "aws_ebs_volume":
	// 	return aws.NewEbsVolume(address, awsRegion, rawValues)
	// case "aws_ebs_snapshot":
	// 	return aws.NewEbsSnapshot(address, awsRegion, rawValues)
	// case "aws_ebs_snapshot_copy":
	// 	return aws.NewEbsSnapshotCopy(address, awsRegion, rawValues)
	// case "aws_launch_configuration":
	// 	return resource.NewBaseResource(address, rawValues, false) // has no cost
	// case "aws_launch_template":
	// 	return resource.NewBaseResource(address, rawValues, false) // has no cost
	// case "aws_autoscaling_group":
	// 	return aws.NewEc2AutoscalingGroup(address, awsRegion, rawValues)
	// case "aws_db_instance":
	// 	return aws.NewRdsInstance(address, awsRegion, rawValues)
	// case "aws_elb":
	// 	return aws.NewElb(address, awsRegion, rawValues, true) // is classic
	// case "aws_lb":
	// 	return aws.NewElb(address, awsRegion, rawValues, false)
	// case "aws_alb": // alias for aws_lb
	// 	return aws.NewElb(address, awsRegion, rawValues, false)
	case "aws_nat_gateway":
		return aws.NewNatGateway(address, awsRegion, rawValues)
		// case "aws_dynamodb_table":
		// 	return aws.NewDynamoDBTable(address, awsRegion, rawValues)
	}
	return nil
}

func ParsePlanJSON(j []byte) ([]resource.Resource, error) {
	planJSON := gjson.ParseBytes(j)
	providerConfig := planJSON.Get("configuration.provider_config")
	plannedValuesJSON := planJSON.Get("planned_values.root_module")
	configurationJSON := planJSON.Get("configuration.root_module")

	resources := make([]resource.Resource, 0)

	resourceMap, err := generateResourceMap(planJSON, providerConfig, plannedValuesJSON)
	if err != nil {
		return resources, err
	}

	err = addReferences(resourceMap, configurationJSON)
	if err != nil {
		return resources, err
	}

	for _, r := range resourceMap {
		resources = append(resources, r)
	}

	// TODO
	// err = setupCostComponents(resources)
	return resources, err
}

func generateResourceMap(planJSON gjson.Result, providerConfig gjson.Result, plannedValuesJSON gjson.Result) (map[string]resource.Resource, error) {
	resourceMap := make(map[string]resource.Resource)
	terraformResources := plannedValuesJSON.Get("resources").Array()

	// Find and create all resources in this module and store in a map
	for _, terraformResource := range terraformResources {
		address := terraformResource.Get("address").String()
		resourceType := terraformResource.Get("type").String()
		var rawValues map[string]interface{}
		if terraformResource.Get("values").Value() != nil {
			rawValues = terraformResource.Get("values").Value().(map[string]interface{})
		} else {
			rawValues = make(map[string]interface{})
		}
		r := createResource(resourceType, address, rawValues, providerConfig)
		if r != nil {
			resourceMap[address] = r
		}
	}

	// Recursively add any resources for child modules
	for _, pvJSON := range plannedValuesJSON.Get("child_modules").Array() {
		moduleResources, err := generateResourceMap(planJSON, providerConfig, pvJSON)
		if err != nil {
			return resourceMap, err
		}
		for address, r := range moduleResources {
			resourceMap[address] = r
		}
	}
	return resourceMap, nil
}

func addReferences(resourceMap map[string]resource.Resource, configurationJSON gjson.Result) error {
	for address, r := range resourceMap {
		resourceConfigJSON := getConfigurationJSONForResourceAddress(configurationJSON, address)

		var refAddressesMap = make(map[string][]string)
		for attribute, attributeJSON := range resourceConfigJSON.Get("expressions").Map() {
			getReferenceAddresses(r, attribute, attributeJSON, &refAddressesMap)
		}

		for attribute, refAddresses := range refAddressesMap {
			for _, refAddress := range refAddresses {
				fullRefAddress := fmt.Sprintf("%s.%s", addressModulePart(address), refAddress)
				if refResource, ok := resourceMap[fullRefAddress]; ok {
					r.AddReference(attribute, refResource)
				}
			}
		}
	}
	return nil
}

func getReferenceAddresses(r resource.Resource, attribute string, attributeJSON gjson.Result, refAddressesMap *map[string][]string) {
	if attributeJSON.Get("references").Exists() {
		for _, ref := range attributeJSON.Get("references").Array() {
			if _, ok := (*refAddressesMap)[attribute]; !ok {
				(*refAddressesMap)[attribute] = make([]string, 0, 1)
			}
			(*refAddressesMap)[attribute] = append((*refAddressesMap)[attribute], ref.String())
		}
	} else if attributeJSON.IsArray() {
		for i, attributeJSONItem := range attributeJSON.Array() {
			getReferenceAddresses(r, fmt.Sprintf("%s.%d", attribute, i), attributeJSONItem, refAddressesMap)
		}
	} else if attributeJSON.Type.String() == "JSON" {
		attributeJSON.ForEach(func(childAttribute gjson.Result, childAttributeJSON gjson.Result) bool {
			getReferenceAddresses(r, fmt.Sprintf("%s.%s", attribute, childAttribute), childAttributeJSON, refAddressesMap)
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
