package aws

import (
	"fmt"
	"infracost/pkg/base"
)

func NewEc2LaunchTemplate(address string, region string, rawValues map[string]interface{}, hasCost bool) base.Resource {
	r := base.NewBaseResource(address, rawValues, hasCost)

	instanceType := rawValues["instance_type"].(string)

	hours := base.NewBasePriceComponent(fmt.Sprintf("instance hours (%s)", instanceType), r, "hour", "hour")
	hours.AddFilters(regionFilters(region))
	hours.AddFilters([]base.Filter{
		{Key: "servicecode", Value: "AmazonEC2"},
		{Key: "productFamily", Value: "Compute Instance"},
		{Key: "operatingSystem", Value: "Linux"},
		{Key: "preInstalledSw", Value: "NA"},
		{Key: "capacitystatus", Value: "Used"},
		{Key: "tenancy", Value: "Shared"},
		{Key: "instanceType", Value: instanceType},
	})
	r.AddPriceComponent(hours)

	rootBlockDeviceRawValues := make(map[string]interface{})
	if r := base.ToGJSON(rawValues).Get("root_block_device.0"); r.Exists() {
		rootBlockDeviceRawValues = r.Value().(map[string]interface{})
	}
	rootBlockDeviceAddress := fmt.Sprintf("%s.root_block_device", address)
	rootBlockDevice := newEc2BlockDevice(rootBlockDeviceAddress, region, rootBlockDeviceRawValues)
	r.AddSubResource(rootBlockDevice)

	blockDeviceMappingsRawValues := make([]interface{}, 0)
	if r := base.ToGJSON(rawValues).Get("block_device_mappings.#.ebs|@flatten"); r.Exists() {
		blockDeviceMappingsRawValues = r.Value().([]interface{})
	}
	for i, blockDeviceMappingRawValues := range blockDeviceMappingsRawValues {
		blockDeviceMappingAddress := fmt.Sprintf("%s.block_device_mappings[%d]", r.Address(), i)
		blockDeviceMapping := newEc2BlockDevice(blockDeviceMappingAddress, region, blockDeviceMappingRawValues.(map[string]interface{}))
		r.AddSubResource(blockDeviceMapping)
	}

	return r
}
