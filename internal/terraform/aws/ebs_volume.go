package aws

import (
	"infracost/pkg/resource"

	"github.com/shopspring/decimal"
)

func ebsVolumeGbQuantity(resource resource.Resource) decimal.Decimal {
	quantity := decimal.NewFromInt(int64(DefaultVolumeSize))

	sizeVal := resource.RawValues()["size"]
	if sizeVal != nil {
		quantity = decimal.NewFromFloat(sizeVal.(float64))
	}

	return quantity
}

func ebsVolumeIopsQuantity(resource resource.Resource) decimal.Decimal {
	quantity := decimal.Zero

	iopsVal := resource.RawValues()["iops"]
	if iopsVal != nil {
		quantity = decimal.NewFromFloat(iopsVal.(float64))
	}

	return quantity
}

func NewEbsVolume(address string, region string, rawValues map[string]interface{}) resource.Resource {
	r := resource.NewBaseResource(address, rawValues, true)

	volumeApiName := "gp2"
	if rawValues["type"] != nil {
		volumeApiName = rawValues["type"].(string)
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
	gb.SetQuantityMultiplierFunc(ebsVolumeGbQuantity)
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
		iops.SetQuantityMultiplierFunc(ebsVolumeIopsQuantity)
		r.AddPriceComponent(iops)
	}

	return r
}
