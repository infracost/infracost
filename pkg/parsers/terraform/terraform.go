package terraform

import (
	"fmt"
	"io/ioutil"
	"os"

	"plancosts/internal/aws"
	"plancosts/pkg/base"

	"github.com/tidwall/gjson"
)

func GetProviderFilters(provider string, region string) []base.Filter {
	switch provider {
	case "aws":
		return aws.GetDefaultFilters(region)
	}
	return []base.Filter{}
}

func GetResourceMapping(resourceType string) *base.ResourceMapping {
	switch resourceType {
	case "aws_instance":
		return aws.Ec2Instance
	case "aws_ebs_volume":
		return aws.EbsVolume
	}
	return nil
}

func ParsePlanFile(filePath string) ([]base.Resource, error) {
	resourceMap := make(map[string]base.Resource, 0)

	planFile, err := os.Open(filePath)
	if err != nil {
		return []base.Resource{}, err
	}
	defer planFile.Close()
	planFileBytes, _ := ioutil.ReadAll(planFile)

	terraformRegion := gjson.GetBytes(planFileBytes, "configuration.provider_config.aws.expressions.region.constant_value").String()
	providerFilters := GetProviderFilters("aws", terraformRegion)

	terraformResources := gjson.GetBytes(planFileBytes, "planned_values.root_module.resources")
	for _, terraformResource := range terraformResources.Array() {
		address := terraformResource.Get("address").String()
		resourceType := terraformResource.Get("type").String()
		rawValues := terraformResource.Get("values").Value().(map[string]interface{})
		if resourceMapping := GetResourceMapping(resourceType); resourceMapping != nil {
			resource := base.NewBaseResource(address, rawValues, resourceMapping, providerFilters)
			resourceMap[address] = resource
		}
	}

	for _, resource := range resourceMap {
		query := fmt.Sprintf(`configuration.root_module.resources.#(address="%s")`, resource.Address())
		terraformResourceConfig := gjson.GetBytes(planFileBytes, query)
		addReferences(resource, terraformResourceConfig, resourceMap)
	}

	resources := make([]base.Resource, 0, len(resourceMap))
	for _, resource := range resourceMap {
		resources = append(resources, resource)
	}

	return resources, nil
}

func addReferences(r base.Resource, resourceConfig gjson.Result, resourceMap map[string]base.Resource) {
	gjson.Get(resourceConfig.String(), "expressions").ForEach(func(key gjson.Result, value gjson.Result) bool {
		var refAddr string
		if value.Get("references").Exists() {
			refAddr = value.Get("references").Array()[0].String()
		} else if len(value.Array()) > 0 {
			idVal := value.Array()[0].Get("id")
			if idVal.Get("references").Exists() {
				refAddr = idVal.Get("references").Array()[0].String()
			}
		}
		if resource, ok := resourceMap[refAddr]; ok {
			r.AddReferences(refAddr, &resource)
		}
		return true
	})
}
