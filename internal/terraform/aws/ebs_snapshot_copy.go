package aws

import (
	"infracost/pkg/base"

	"github.com/shopspring/decimal"
)

func ebsSnapshotCopyGbQuantity(resource base.Resource) decimal.Decimal {
	quantity := decimal.NewFromInt(int64(DefaultVolumeSize))

	sourceSnapshot := resource.References()["source_snapshot_id"]
	if sourceSnapshot == nil {
		return quantity
	}

	volume := sourceSnapshot.References()["volume_id"]
	if volume == nil {
		return quantity
	}

	sizeVal := volume.RawValues()["size"]
	if sizeVal != nil {
		quantity = decimal.NewFromFloat(sizeVal.(float64))
	}
	return quantity
}

func NewEbsSnapshotCopy(address string, region string, rawValues map[string]interface{}) base.Resource {
	r := base.NewBaseResource(address, rawValues, true)

	gb := base.NewBasePriceComponent("GB", r, "GB/month", "month")
	gb.AddFilters(regionFilters(region))
	gb.AddFilters([]base.Filter{
		{Key: "servicecode", Value: "AmazonEC2"},
		{Key: "productFamily", Value: "Storage Snapshot"},
		{Key: "usagetype", Value: "/EBS:SnapshotUsage$/", Operation: "REGEX"},
	})
	gb.SetQuantityMultiplierFunc(ebsSnapshotCopyGbQuantity)
	r.AddPriceComponent(gb)

	return r
}
