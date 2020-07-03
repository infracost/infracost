package aws

import (
	"infracost/pkg/base"

	"github.com/shopspring/decimal"
)

type EbsSnapshotCopyGB struct {
	*BaseAwsPriceComponent
}

func NewEbsSnapshotCopyGB(name string, resource *EbsSnapshotCopy) *EbsSnapshotCopyGB {
	c := &EbsSnapshotCopyGB{
		NewBaseAwsPriceComponent(name, resource.BaseAwsResource, "month"),
	}

	c.defaultFilters = []base.Filter{
		{Key: "servicecode", Value: "AmazonEC2"},
		{Key: "productFamily", Value: "Storage Snapshot"},
		{Key: "usagetype", Value: "/EBS:SnapshotUsage$/", Operation: "REGEX"},
	}

	return c
}

func (c *EbsSnapshotCopyGB) HourlyCost() decimal.Decimal {
	hourlyCost := c.BaseAwsPriceComponent.HourlyCost()
	size := decimal.NewFromInt(int64(DefaultVolumeSize))
	volumeGBPriceComponent := base.GetPriceComponent(c.AwsResource().References()["source_snapshot_id"].References()["volume_id"], "GB")
	if volumeGBPriceComponent == nil {
		sizeVal := volumeGBPriceComponent.(*EbsVolumeGB).AwsResource().RawValues()["size"]
		if sizeVal != nil {
			size = decimal.NewFromFloat(sizeVal.(float64))
		}
	}
	return hourlyCost.Mul(size)
}

type EbsSnapshotCopy struct {
	*BaseAwsResource
}

func NewEbsSnapshotCopy(address string, region string, rawValues map[string]interface{}) *EbsSnapshotCopy {
	r := &EbsSnapshotCopy{
		BaseAwsResource: NewBaseAwsResource(address, region, rawValues),
	}
	r.BaseAwsResource.priceComponents = []base.PriceComponent{
		NewEbsSnapshotCopyGB("GB", r),
	}
	return r
}
