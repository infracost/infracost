package aws

import (
	"fmt"
	"infracost/pkg/resource"
	"math"

	"github.com/shopspring/decimal"
)

func ec2LaunchTemplateHoursQuantityFactory(r resource.Resource, purchaseOption string, onDemandBaseCount int, onDemandPerc int) func(resource resource.Resource) decimal.Decimal {
	return func(r resource.Resource) decimal.Decimal {
		onDemandCount := onDemandBaseCount
		remainingCount := r.ResourceCount() - onDemandCount
		onDemandCount += int(math.Ceil(float64(remainingCount) * float64(onDemandPerc) / float64(100)))
		spotCount := r.ResourceCount() - onDemandCount

		purchaseOptionCount := onDemandCount
		if purchaseOption == "spot" {
			purchaseOptionCount = spotCount
		}
		return decimal.NewFromInt(int64(purchaseOptionCount)).Div(decimal.NewFromInt(int64(r.ResourceCount())))
	}
}

func NewEc2LaunchTemplate(address string, region string, rawValues map[string]interface{}, onDemandBaseCount int, onDemandPerc int) resource.Resource {
	r := resource.NewBaseResource(address, rawValues, true)

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

	if onDemandBaseCount > 0 || onDemandPerc > 0 {
		onDemandHoursPriceFilter := &resource.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		}
		onDemandHours := resource.NewBasePriceComponent(fmt.Sprintf("instance hours (%s)", instanceType), r, "hour", "hour", hoursProductFilter, onDemandHoursPriceFilter)
		onDemandHours.SetQuantityMultiplierFunc(ec2LaunchTemplateHoursQuantityFactory(r, "on_demand", onDemandBaseCount, onDemandPerc))
		r.AddPriceComponent(onDemandHours)
	}

	if onDemandPerc != 100 {
		spotHoursPriceFilter := &resource.PriceFilter{
			PurchaseOption: strPtr("spot"),
		}
		spotHours := resource.NewBasePriceComponent(fmt.Sprintf("instance hours (%s, spot)", instanceType), r, "hour", "hour", hoursProductFilter, spotHoursPriceFilter)
		spotHours.SetQuantityMultiplierFunc(ec2LaunchTemplateHoursQuantityFactory(r, "spot", onDemandBaseCount, onDemandPerc))
		r.AddPriceComponent(spotHours)
	}

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
