package aws

import (
	"fmt"
	"infracost/pkg/resource"

	"github.com/shopspring/decimal"
)

func ec2BlockDeviceGbQuantity(resource resource.Resource) decimal.Decimal {
	quantity := decimal.NewFromInt(int64(DefaultVolumeSize))

	sizeVal := resource.RawValues()["volume_size"]
	if sizeVal != nil {
		quantity = decimal.NewFromFloat(sizeVal.(float64))
	}

	return quantity
}

func ec2BlockDeviceIopsQuantity(resource resource.Resource) decimal.Decimal {
	quantity := decimal.Zero

	iopsVal := resource.RawValues()["iops"]
	if iopsVal != nil {
		quantity = decimal.NewFromFloat(iopsVal.(float64))
	}

	return quantity
}

func newEc2BlockDevice(address string, region string, rawValues map[string]interface{}) resource.Resource {
	r := resource.NewBaseResource(address, rawValues, true)

	volumeApiName := "gp2"
	if rawValues["volume_type"] != nil {
		volumeApiName = rawValues["volume_type"].(string)
	}

	gbProductFilter := &resource.ProductFilter{
		VendorName:    strPtr("aws"),
		Region:        strPtr(region),
		Service:       strPtr("AmazonEC2"),
		ProductFamily: strPtr("Storage"),
		AttributeFilters: &[]resource.AttributeFilter{
			{Key: "volumeApiName", Value: strPtr(volumeApiName)},
		},
	}
	gb := resource.NewBasePriceComponent("GB", r, "GB/month", "month", gbProductFilter, nil)
	gb.SetQuantityMultiplierFunc(ec2BlockDeviceGbQuantity)
	r.AddPriceComponent(gb)

	if volumeApiName == "io1" {
		iopsProductFilter := &resource.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonEC2"),
			ProductFamily: strPtr("System Operation"),
			AttributeFilters: &[]resource.AttributeFilter{
				{Key: "volumeApiName", Value: strPtr(volumeApiName)},
				{Key: "usagetype", ValueRegex: strPtr("/EBS:VolumeP-IOPS.piops/")},
			},
		}
		iops := resource.NewBasePriceComponent("IOPS", r, "IOPS/month", "month", iopsProductFilter, nil)
		iops.SetQuantityMultiplierFunc(ec2BlockDeviceIopsQuantity)
		r.AddPriceComponent(iops)
	}

	return r
}

func NewEc2Instance(address string, region string, rawValues map[string]interface{}) resource.Resource {
	r := resource.NewBaseResource(address, rawValues, true)

	instanceType := rawValues["instance_type"].(string)
	tenancy := "Shared"
	if rawValues["tenancy"] != nil && rawValues["tenancy"].(string) == "dedicated" {
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

	if r.RawValues()["ebs_block_device"] != nil {
		ebsBlockDevicesRawValues := rawValues["ebs_block_device"].([]interface{})
		for i, ebsBlockDeviceRawValues := range ebsBlockDevicesRawValues {
			ebsBlockDeviceAddress := fmt.Sprintf("%s.ebs_block_device[%d]", r.Address(), i)
			ebsBlockDevice := newEc2BlockDevice(ebsBlockDeviceAddress, region, ebsBlockDeviceRawValues.(map[string]interface{}))
			r.AddSubResource(ebsBlockDevice)
		}
	}

	return r
}
