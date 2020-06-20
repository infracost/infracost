package terraform

import (
	"fmt"
	"io/ioutil"
	"os"

	"plancosts/internal/aws_terraform"
	"plancosts/pkg/base"

	"github.com/tidwall/gjson"
)

func createResource(resourceType string, address string, rawValues map[string]interface{}, providerConfig gjson.Result) base.Resource {
	awsRegion := providerConfig.Get("aws.expressions.region.constant_value").String()

	switch resourceType {
	case "aws_instance":
		return aws_terraform.NewEc2Instance(address, awsRegion, rawValues)
	case "aws_ebs_volume":
		return aws_terraform.NewEbsVolume(address, awsRegion, rawValues)
	case "aws_ebs_snapshot":
		return aws_terraform.NewEbsSnapshot(address, awsRegion, rawValues)
	case "aws_ebs_snapshot_copy":
		return aws_terraform.NewEbsSnapshotCopy(address, awsRegion, rawValues)
	case "aws_launch_configuration":
		return aws_terraform.NewEc2LaunchConfiguration(address, awsRegion, rawValues)
	case "aws_launch_template":
		return aws_terraform.NewEc2LaunchTemplate(address, awsRegion, rawValues)
	case "aws_autoscaling_group":
		return aws_terraform.NewEc2AutoscalingGroup(address, awsRegion, rawValues)
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

	providerConfig := gjson.GetBytes(planFileBytes, "configuration.provider_config")
	terraformResources := gjson.GetBytes(planFileBytes, "planned_values.root_module.resources")
	for _, terraformResource := range terraformResources.Array() {
		address := terraformResource.Get("address").String()
		resourceType := terraformResource.Get("type").String()
		rawValues := terraformResource.Get("values").Value().(map[string]interface{})
		resource := createResource(resourceType, address, rawValues, providerConfig)
		if resource != nil {
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
			r.AddReference(key.String(), resource)
		}
		return true
	})
}
