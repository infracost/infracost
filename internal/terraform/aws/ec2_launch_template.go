package aws

import (
	"fmt"
	"infracost/pkg/resource"
)

func NewEc2LaunchTemplate(address string, region string, rawValues map[string]interface{}, hasCost bool) resource.Resource {
	r := resource.NewBaseResource(address, rawValues, hasCost)

	instanceType := rawValues["instance_type"].(string)

	hoursProductFilter := &resource.ProductFilter{
		VendorName:    strPtr("aws"),
		Region:        strPtr(region),
		Service:       strPtr("AmazonEC2"),
		ProductFamily: strPtr("Compute Instance"),
		AttributeFilters: &[]resource.AttributeFilter{
			{Key: "instanceType", Value: strPtr(instanceType)},
			{Key: "tenancy", Value: strPtr("Shared")},
			{Key: "operatingSystem", Value: strPtr("Linux")},
			{Key: "preInstalledSw", Value: strPtr("NA")},
			{Key: "capacitystatus", Value: strPtr("Used")},
		},
	}
	hoursPriceFilter := &resource.PriceFilter{
		PurchaseOption: strPtr("on_demand"),
	}
	hours := resource.NewBasePriceComponent(fmt.Sprintf("instance hours (%s)", instanceType), r, "hour", "hour", hoursProductFilter, hoursPriceFilter)
	r.AddPriceComponent(hours)

	rootBlockDeviceRawValues := make(map[string]interface{})
	if r := resource.ToGJSON(rawValues).Get("root_block_device.0"); r.Exists() {
		rootBlockDeviceRawValues = r.Value().(map[string]interface{})
	}
	rootBlockDeviceAddress := fmt.Sprintf("%s.root_block_device", address)
	rootBlockDevice := newEc2BlockDevice(rootBlockDeviceAddress, region, rootBlockDeviceRawValues)
	r.AddSubResource(rootBlockDevice)

	blockDeviceMappingsRawValues := make([]interface{}, 0)
	if r := resource.ToGJSON(rawValues).Get("block_device_mappings.#.ebs|@flatten"); r.Exists() {
		blockDeviceMappingsRawValues = r.Value().([]interface{})
	}
	for i, blockDeviceMappingRawValues := range blockDeviceMappingsRawValues {
		blockDeviceMappingAddress := fmt.Sprintf("%s.block_device_mappings[%d]", r.Address(), i)
		blockDeviceMapping := newEc2BlockDevice(blockDeviceMappingAddress, region, blockDeviceMappingRawValues.(map[string]interface{}))
		r.AddSubResource(blockDeviceMapping)
	}

	return r
}
