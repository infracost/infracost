package aws_terraform

import (
	"fmt"
	"plancosts/pkg/base"
)

type Ec2LaunchTemplateHours struct {
	*BaseAwsPriceComponent
}

func NewEc2LaunchTemplateHours(name string, resource *Ec2LaunchTemplate) *Ec2LaunchTemplateHours {
	c := &Ec2LaunchTemplateHours{
		NewBaseAwsPriceComponent(name, resource.BaseAwsResource, "hour"),
	}

	c.defaultFilters = []base.Filter{
		{Key: "servicecode", Value: "AmazonEC2"},
		{Key: "productFamily", Value: "Compute Instance"},
		{Key: "operatingSystem", Value: "Linux"},
		{Key: "preInstalledSw", Value: "NA"},
		{Key: "capacitystatus", Value: "Used"},
		{Key: "tenancy", Value: "Shared"},
	}

	c.valueMappings = []base.ValueMapping{
		{FromKey: "instance_type", ToKey: "instanceType"},
	}

	return c
}

type Ec2LaunchTemplate struct {
	*BaseAwsResource
}

func NewEc2LaunchTemplate(address string, region string, rawValues map[string]interface{}) *Ec2LaunchTemplate {
	r := &Ec2LaunchTemplate{
		NewBaseAwsResource(address, region, rawValues),
	}
	r.BaseAwsResource.priceComponents = []base.PriceComponent{
		NewEc2LaunchTemplateHours("Instance hours", r),
	}

	subResources := make([]base.Resource, 0)
	block_device_mappings := r.rawValues["block_device_mappings"]
	if block_device_mappings != nil {
		for i, block_device_mapping := range block_device_mappings.([]interface{}) {
			address := fmt.Sprintf("%s.block_device_mappings[%d]", r.Address(), i)
			rawValues := block_device_mapping.(map[string]interface{})["ebs"].([]interface{})[0].(map[string]interface{})
			subResources = append(subResources, NewEc2BlockDevice(address, r.region, rawValues))
		}
	}
	r.BaseAwsResource.subResources = subResources

	return r
}
func (r *Ec2LaunchTemplate) HasCost() bool {
	return false
}
