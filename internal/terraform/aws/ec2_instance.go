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

	gb := resource.NewBasePriceComponent("GB", r, "GB/month", "month")
	gb.AddFilters(regionFilters(region))
	gb.AddFilters([]resource.Filter{
		{Key: "servicecode", Value: "AmazonEC2"},
		{Key: "productFamily", Value: "Storage"},
		{Key: "volumeApiName", Value: volumeApiName},
	})
	gb.SetQuantityMultiplierFunc(ec2BlockDeviceGbQuantity)
	r.AddPriceComponent(gb)

	if volumeApiName == "io1" {
		iops := resource.NewBasePriceComponent("IOPS", r, "IOPS/month", "month")
		iops.AddFilters(regionFilters(region))
		iops.AddFilters([]resource.Filter{
			{Key: "servicecode", Value: "AmazonEC2"},
			{Key: "productFamily", Value: "System Operation"},
			{Key: "usagetype", Value: "/EBS:VolumeP-IOPS.piops/", Operation: "REGEX"},
			{Key: "volumeApiName", Value: volumeApiName},
		})
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
