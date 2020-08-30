package aws

import (
	"fmt"
	"infracost/pkg/schema"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func AwsInstance(d *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()
	instanceType := d.Get("instance_type").String()

	tenancy := "Shared"
	if d.Get("tenancy").Exists() && d.Get("tenancy").String() == "dedicated" {
		tenancy = "Dedicated"
	}

	subResources := make([]*schema.Resource, 0)
	subResources = append(subResources, rootBlockDevice(d.Get("root_block_device"), region))
	for _, ebsBlockDevice := range ebsBlockDevices(d.Get("ebs_block_device"), region) {
		subResources = append(subResources, ebsBlockDevice)
	}

	return &schema.Resource{
		Name:         d.Address,
		SubResources: subResources,
		CostComponents: []*schema.CostComponent{
			{
				Name:           "Instance hours",
				Unit:           "hours",
				HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonEC2"),
					ProductFamily: strPtr("Compute Instance"),
					AttributeFilters: &[]schema.AttributeFilter{
						{Key: "instanceType", Value: strPtr(instanceType)},
						{Key: "tenancy", Value: strPtr(tenancy)},
						{Key: "operatingSystem", Value: strPtr("Linux")},
						{Key: "preInstalledSw", Value: strPtr("NA")},
						{Key: "capacitystatus", Value: strPtr("Used")},
					},
				},
				PriceFilter: &schema.PriceFilter{
					PurchaseOption: strPtr("on_demand"),
				},
			},
		},
	}
}

func rootBlockDevice(d gjson.Result, region string) *schema.Resource {
	return ebsBlockDevice("root_block_device", d, region)
}

func ebsBlockDevices(d gjson.Result, region string) []*schema.Resource {
	resources := make([]*schema.Resource, 0)
	for i, data := range d.Array() {
		name := fmt.Sprintf("ebs_block_device[%d]", i)
		resources = append(resources, ebsBlockDevice(name, data, region))
	}
	return resources
}

func ebsBlockDevice(name string, d gjson.Result, region string) *schema.Resource {
	volumeApiName := "gp2"
	if d.Get("volume_type").Exists() {
		volumeApiName = d.Get("volume_type").String()
	}

	gbVal := decimal.NewFromInt(int64(defaultVolumeSize))
	if d.Get("volume_size").Exists() {
		gbVal = decimal.NewFromFloat(d.Get("volume_size").Float())
	}

	iopsVal := decimal.NewFromInt(int64(defaultVolumeSize))
	if d.Get("iops").Exists() {
		iopsVal = decimal.NewFromFloat(d.Get("iops").Float())
	}

	costComponents := []*schema.CostComponent{
		{
			Name:            "Storage",
			Unit:            "GB/months",
			MonthlyQuantity: &gbVal,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonEC2"),
				ProductFamily: strPtr("Storage"),
				AttributeFilters: &[]schema.AttributeFilter{
					{Key: "volumeApiName", Value: strPtr(volumeApiName)},
				},
			},
		},
	}

	if volumeApiName == "io1" {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "IOPS",
			Unit:            "IOPS/months",
			MonthlyQuantity: &iopsVal,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonEC2"),
				ProductFamily: strPtr("System Operation"),
				AttributeFilters: &[]schema.AttributeFilter{
					{Key: "volumeApiName", Value: strPtr(volumeApiName)},
					{Key: "usagetype", ValueRegex: strPtr("/EBS:VolumeP-IOPS.piops/")},
				},
			},
		})
	}

	return &schema.Resource{
		Name:           name,
		CostComponents: costComponents,
	}
}
