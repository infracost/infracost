package aws

import (
	"fmt"
	"infracost/pkg/resource"
)

func NewEc2LaunchConfiguration(address string, region string, rawValues map[string]interface{}, hasCost bool) resource.Resource {
	r := resource.NewBaseResource(address, rawValues, hasCost)

	instanceType := rawValues["instance_type"].(string)
	tenancy := "Shared"
	if rawValues["placement_tenancy"] != nil && rawValues["placement_tenancy"].(string) == "dedicated" {
		tenancy = "Dedicated"
	}

	hours := resource.NewBasePriceComponent(fmt.Sprintf("instance hours (%s)", instanceType), r, "hour", "hour")
	hours.AddFilters(regionFilters(region))
	hours.AddFilters([]resource.Filter{
		{Key: "servicecode", Value: "AmazonEC2"},
		{Key: "productFamily", Value: "Compute Instance"},
		{Key: "operatingSystem", Value: "Linux"},
		{Key: "preInstalledSw", Value: "NA"},
		{Key: "capacitystatus", Value: "Used"},
		{Key: "tenancy", Value: tenancy},
		{Key: "instanceType", Value: instanceType},
	})
	r.AddPriceComponent(hours)

	rootBlockDeviceRawValues := make(map[string]interface{})
	if r := resource.ToGJSON(rawValues).Get("root_block_device.0"); r.Exists() {
		rootBlockDeviceRawValues = r.Value().(map[string]interface{})
	}
	rootBlockDeviceAddress := fmt.Sprintf("%s.root_block_device", address)
	rootBlockDevice := newEc2BlockDevice(rootBlockDeviceAddress, region, rootBlockDeviceRawValues)
	r.AddSubResource(rootBlockDevice)

	if rawValues["ebs_block_device"] != nil {
		ebsBlockDevicesRawValues := rawValues["ebs_block_device"].([]interface{})
		for i, ebsBlockDeviceRawValues := range ebsBlockDevicesRawValues {
			ebsBlockDeviceAddress := fmt.Sprintf("%s.ebs_block_device[%d]", r.Address(), i)
			ebsBlockDevice := newEc2BlockDevice(ebsBlockDeviceAddress, region, ebsBlockDeviceRawValues.(map[string]interface{}))
			r.AddSubResource(ebsBlockDevice)
		}
	}

	return r
}
