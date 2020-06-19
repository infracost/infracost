package aws

import (
	"fmt"
	"plancosts/pkg/base"

	"github.com/shopspring/decimal"
)

var EbsVolumeGB = &base.PriceMapping{
	TimeUnit: "month",

	DefaultFilters: []base.Filter{
		{Key: "servicecode", Value: "AmazonEC2"},
		{Key: "productFamily", Value: "Storage"},
		{Key: "volumeApiName", Value: "gp2"},
	},

	ValueMappings: []base.ValueMapping{
		{FromKey: "type", ToKey: "volumeApiName"},
	},

	CalculateCost: func(price decimal.Decimal, values map[string]interface{}) decimal.Decimal {
		size := decimal.NewFromInt(int64(8))
		if sizeStr, ok := values["size"]; ok {
			size, _ = decimal.NewFromString(fmt.Sprintf("%v", sizeStr))
		}
		return price.Mul(size)
	},
}

var EbsVolumeIOPS = &base.PriceMapping{
	TimeUnit: "month",

	DefaultFilters: []base.Filter{
		{Key: "servicecode", Value: "AmazonEC2"},
		{Key: "productFamily", Value: "System Operation"},
		{Key: "usagetype", Value: "/EBS:VolumeP-IOPS.piops/", Operation: "REGEX"},
		{Key: "volumeApiName", Value: "gp2"},
	},

	ValueMappings: []base.ValueMapping{
		{FromKey: "type", ToKey: "volumeApiName"},
	},

	ShouldSkip: func(values map[string]interface{}) bool {
		return values["type"] != "io1"
	},

	CalculateCost: func(price decimal.Decimal, values map[string]interface{}) decimal.Decimal {
		iops, _ := decimal.NewFromString(fmt.Sprintf("%v", values["iops"]))
		return price.Mul(iops)
	},
}

var EbsVolume = &base.ResourceMapping{
	PriceMappings: map[string]*base.PriceMapping{
		"GB":   EbsVolumeGB,
		"IOPS": EbsVolumeIOPS,
	},
}
