package aws

import (
	"infracost/pkg/base"

	"github.com/shopspring/decimal"
)

func ebsVolumeGbQuantity(resource base.Resource) decimal.Decimal {
	quantity := decimal.NewFromInt(int64(DefaultVolumeSize))

	sizeVal := resource.RawValues()["size"]
	if sizeVal != nil {
		quantity = decimal.NewFromFloat(sizeVal.(float64))
	}

	return quantity
}

func ebsVolumeIopsQuantity(resource base.Resource) decimal.Decimal {
	quantity := decimal.Zero

	iopsVal := resource.RawValues()["iops"]
	if iopsVal != nil {
		quantity = decimal.NewFromFloat(iopsVal.(float64))
	}

	return quantity
}

func NewEbsVolume(address string, region string, rawValues map[string]interface{}) base.Resource {
	r := base.NewBaseResource(address, rawValues, true)

	volumeApiName := "gp2"
	if rawValues["type"] != nil {
		volumeApiName = rawValues["type"].(string)
	}

	gb := base.NewBasePriceComponent("GB", r, "GB/month", "month")
	gb.AddFilters(regionFilters(region))
	gb.AddFilters([]base.Filter{
		{Key: "servicecode", Value: "AmazonEC2"},
		{Key: "productFamily", Value: "Storage"},
		{Key: "volumeApiName", Value: volumeApiName},
	})
	gb.SetQuantityMultiplierFunc(ebsVolumeGbQuantity)
	r.AddPriceComponent(gb)

	if volumeApiName == "io1" {
		iops := base.NewBasePriceComponent("IOPS", r, "IOPS/month", "month")
		iops.AddFilters(regionFilters(region))
		iops.AddFilters([]base.Filter{
			{Key: "servicecode", Value: "AmazonEC2"},
			{Key: "productFamily", Value: "System Operation"},
			{Key: "usagetype", Value: "/EBS:VolumeP-IOPS.piops/", Operation: "REGEX"},
			{Key: "volumeApiName", Value: volumeApiName},
		})
		iops.SetQuantityMultiplierFunc(ebsVolumeIopsQuantity)
		r.AddPriceComponent(iops)
	}

	return r
}
