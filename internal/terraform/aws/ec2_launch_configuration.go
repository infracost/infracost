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

	hoursProductFilter := &resource.ProductFilter{
		VendorName:    strPtr("aws"),
		Region:        strPtr(region),
		Service:       strPtr("AmazonEC2"),
		ProductFamily: strPtr("Compute Instance"),
		AttributeFilters: &[]resource.AttributeFilter{
			{Key: "instanceType", Value: strPtr(instanceType)},
			{Key: "tenancy", Value: strPtr(tenancy)},
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
