package aws

import (
	"infracost/pkg/resource"

	"github.com/shopspring/decimal"
)

func ebsSnapshotCopyGbQuantity(resource resource.Resource) decimal.Decimal {
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

func NewEbsSnapshotCopy(address string, region string, rawValues map[string]interface{}) resource.Resource {
	r := resource.NewBaseResource(address, rawValues, true)

	gb := resource.NewBasePriceComponent("GB", r, "GB/month", "month")
	gb.AddFilters(regionFilters(region))
	gb.AddFilters([]resource.Filter{
		{Key: "servicecode", Value: "AmazonEC2"},
		{Key: "productFamily", Value: "Storage Snapshot"},
		{Key: "usagetype", Value: "/EBS:SnapshotUsage$/", Operation: "REGEX"},
	})
	gb.SetQuantityMultiplierFunc(ebsSnapshotCopyGbQuantity)
	r.AddPriceComponent(gb)

	return r
}
