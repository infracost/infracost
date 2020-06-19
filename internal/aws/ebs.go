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

	CalculateCost: func(price decimal.Decimal, resource base.Resource) decimal.Decimal {
		size := decimal.NewFromInt(int64(8))
		if sizeStr, ok := resource.RawValues()["size"]; ok {
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

	CalculateCost: func(price decimal.Decimal, resource base.Resource) decimal.Decimal {
		iops, _ := decimal.NewFromString(fmt.Sprintf("%v", resource.RawValues()["iops"]))
		return price.Mul(iops)
	},
}

var EbsVolume = &base.ResourceMapping{
	PriceMappings: map[string]*base.PriceMapping{
		"GB":   EbsVolumeGB,
		"IOPS": EbsVolumeIOPS,
	},
}

var EbsSnapshotGB = &base.PriceMapping{
	TimeUnit: "month",

	DefaultFilters: []base.Filter{
		{Key: "servicecode", Value: "AmazonEC2"},
		{Key: "productFamily", Value: "Storage Snapshot"},
	},

	CalculateCost: func(price decimal.Decimal, resource base.Resource) decimal.Decimal {
		size := decimal.NewFromInt(int64(8))
		if sizeStr, ok := resource.References()["volume_id"].RawValues()["size"]; ok {
			size, _ = decimal.NewFromString(fmt.Sprintf("%v", sizeStr))
		}
		return price.Mul(size)
	},
}

var EbsSnapshot = &base.ResourceMapping{
	PriceMappings: map[string]*base.PriceMapping{
		"GB": EbsSnapshotGB,
	},
}

var EbsSnapshotCopyGB = &base.PriceMapping{
	TimeUnit:       EbsSnapshotGB.TimeUnit,
	DefaultFilters: EbsSnapshotGB.DefaultFilters,

	CalculateCost: func(price decimal.Decimal, resource base.Resource) decimal.Decimal {
		size := decimal.NewFromInt(int64(8))
		if sizeStr, ok := resource.References()["source_snapshot_id"].References()["volume_id"].RawValues()["size"]; ok {
			size, _ = decimal.NewFromString(fmt.Sprintf("%v", sizeStr))
		}
		return price.Mul(size)
	},
}

var EbsSnapshotCopy = &base.ResourceMapping{
	PriceMappings: map[string]*base.PriceMapping{
		"GB": EbsSnapshotCopyGB,
	},
}
