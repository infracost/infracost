package aws

import (
	"fmt"
	"infracost/pkg/base"

	"github.com/shopspring/decimal"
)

type Ec2BlockDeviceGB struct {
	*BaseAwsPriceComponent
}

func NewEc2BlockDeviceGB(name string, resource *Ec2BlockDevice) *Ec2BlockDeviceGB {
	c := &Ec2BlockDeviceGB{
		NewBaseAwsPriceComponent(name, resource.BaseAwsResource, "month"),
	}

	c.defaultFilters = []base.Filter{
		{Key: "servicecode", Value: "AmazonEC2"},
		{Key: "productFamily", Value: "Storage"},
		{Key: "volumeApiName", Value: "gp2"},
	}

	c.valueMappings = []base.ValueMapping{
		{FromKey: "volume_type", ToKey: "volumeApiName"},
	}

	return c
}

func (c *Ec2BlockDeviceGB) HourlyCost() decimal.Decimal {
	hourlyCost := c.BaseAwsPriceComponent.HourlyCost()
	size := decimal.NewFromInt(int64(DefaultVolumeSize))
	if c.AwsResource().RawValues()["volume_size"] != nil {
		size = decimal.NewFromFloat(c.AwsResource().RawValues()["volume_size"].(float64))
	}
	return hourlyCost.Mul(size)
}

type Ec2BlockDeviceIOPS struct {
	*BaseAwsPriceComponent
}

func NewEc2BlockDeviceIOPS(name string, resource *Ec2BlockDevice) *Ec2BlockDeviceIOPS {
	c := &Ec2BlockDeviceIOPS{
		NewBaseAwsPriceComponent(name, resource.BaseAwsResource, "month"),
	}

	c.defaultFilters = []base.Filter{
		{Key: "servicecode", Value: "AmazonEC2"},
		{Key: "productFamily", Value: "System Operation"},
		{Key: "usagetype", Value: "/EBS:VolumeP-IOPS.piops/", Operation: "REGEX"},
		{Key: "volumeApiName", Value: "gp2"},
	}

	c.valueMappings = []base.ValueMapping{
		{FromKey: "volume_type", ToKey: "volumeApiName"},
	}

	return c
}

func (c *Ec2BlockDeviceIOPS) HourlyCost() decimal.Decimal {
	hourlyCost := c.BaseAwsPriceComponent.HourlyCost()
	iops := decimal.NewFromInt(int64(0))
	if c.AwsResource().RawValues()["iops"] != nil {
		iops = decimal.NewFromFloat(c.AwsResource().RawValues()["iops"].(float64))
	}
	return hourlyCost.Mul(iops)
}

type Ec2BlockDevice struct {
	*BaseAwsResource
}

func NewEc2BlockDevice(address string, region string, rawValues map[string]interface{}) *Ec2BlockDevice {
	r := &Ec2BlockDevice{
		BaseAwsResource: NewBaseAwsResource(address, region, rawValues),
	}
	priceComponents := []base.PriceComponent{
		NewEc2BlockDeviceGB("GB", r),
	}
	if r.RawValues()["volume_type"] == "io1" {
		priceComponents = append(priceComponents, NewEc2BlockDeviceIOPS("IOPS", r))
	}
	r.BaseAwsResource.priceComponents = priceComponents
	return r
}

type Ec2InstanceHours struct {
	*BaseAwsPriceComponent
}

func NewEc2InstanceHours(name string, resource *Ec2Instance) *Ec2InstanceHours {
	c := &Ec2InstanceHours{
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
		{FromKey: "tenancy", ToKey: "tenancy"},
	}

	return c
}

type Ec2Instance struct {
	*BaseAwsResource
}

func NewEc2Instance(address string, region string, rawValues map[string]interface{}) *Ec2Instance {
	r := &Ec2Instance{
		NewBaseAwsResource(address, region, rawValues),
	}
	r.BaseAwsResource.priceComponents = []base.PriceComponent{
		NewEc2InstanceHours("Instance hours", r),
	}

	subResources := make([]base.Resource, 0)
	subResourceAddress := fmt.Sprintf("%s.root_block_device", r.Address())
	if r.RawValues()["root_block_device"] != nil {
		rootBlockDevices := r.RawValues()["root_block_device"].([]interface{})
		subResources = append(subResources, NewEc2BlockDevice(subResourceAddress, r.region, rootBlockDevices[0].(map[string]interface{})))
	} else {
		subResources = append(subResources, NewEc2BlockDevice(subResourceAddress, r.region, make(map[string]interface{})))
	}

	if r.RawValues()["ebs_block_device"] != nil {
		ebsBlockDevices := r.RawValues()["ebs_block_device"].([]interface{})
		for i, ebsBlockDevice := range ebsBlockDevices {
			subResourceAddress := fmt.Sprintf("%s.ebs_block_device[%d]", r.Address(), i)
			subResources = append(subResources, NewEc2BlockDevice(subResourceAddress, r.region, ebsBlockDevice.(map[string]interface{})))
		}
	}

	r.BaseAwsResource.subResources = subResources

	return r
}
