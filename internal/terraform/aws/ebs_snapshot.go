package aws

import (
	"plancosts/pkg/base"

	"github.com/shopspring/decimal"
)

type EbsSnapshotGB struct {
	*BaseAwsPriceComponent
}

func NewEbsSnapshotGB(name string, resource *EbsSnapshot) *EbsSnapshotGB {
	c := &EbsSnapshotGB{
		NewBaseAwsPriceComponent(name, resource.BaseAwsResource, "month"),
	}

	c.defaultFilters = []base.Filter{
		{Key: "servicecode", Value: "AmazonEC2"},
		{Key: "productFamily", Value: "Storage Snapshot"},
	}

	return c
}

func (c *EbsSnapshotGB) HourlyCost() decimal.Decimal {
	hourlyCost := c.BaseAwsPriceComponent.HourlyCost()
	size := decimal.NewFromInt(int64(DefaultVolumeSize))
	volumeGBPriceComponent := base.GetPriceComponent(c.AwsResource().References()["volume_id"], "GB")
	if volumeGBPriceComponent == nil {
		sizeVal := volumeGBPriceComponent.(*EbsVolumeGB).AwsResource().RawValues()["size"]
		if sizeVal != nil {
			size = decimal.NewFromFloat(sizeVal.(float64))
		}
	}
	return hourlyCost.Mul(size)
}

type EbsSnapshot struct {
	*BaseAwsResource
}

func NewEbsSnapshot(address string, region string, rawValues map[string]interface{}) *EbsSnapshot {
	r := &EbsSnapshot{
		BaseAwsResource: NewBaseAwsResource(address, region, rawValues),
	}
	r.BaseAwsResource.priceComponents = []base.PriceComponent{
		NewEbsSnapshotGB("GB", r),
	}
	return r
}
