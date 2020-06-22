package aws

import (
	"plancosts/pkg/base"

	"github.com/shopspring/decimal"
)

type EbsVolumeGB struct {
	*BaseAwsPriceComponent
}

func NewEbsVolumeGB(name string, resource *EbsVolume) *EbsVolumeGB {
	c := &EbsVolumeGB{
		NewBaseAwsPriceComponent(name, resource.BaseAwsResource, "month"),
	}

	c.defaultFilters = []base.Filter{
		{Key: "servicecode", Value: "AmazonEC2"},
		{Key: "productFamily", Value: "Storage"},
		{Key: "volumeApiName", Value: "gp2"},
	}

	c.valueMappings = []base.ValueMapping{
		{FromKey: "type", ToKey: "volumeApiName"},
	}

	return c
}

func (c *EbsVolumeGB) HourlyCost() decimal.Decimal {
	hourlyCost := c.BaseAwsPriceComponent.HourlyCost()
	size := decimal.NewFromInt(int64(DefaultVolumeSize))
	if c.AwsResource().RawValues()["size"] != nil {
		size = decimal.NewFromFloat(c.AwsResource().RawValues()["size"].(float64))
	}
	return hourlyCost.Mul(size)
}

type EbsVolumeIOPS struct {
	*BaseAwsPriceComponent
}

func NewEbsVolumeIOPS(name string, resource *EbsVolume) *EbsVolumeIOPS {
	c := &EbsVolumeIOPS{
		NewBaseAwsPriceComponent(name, resource.BaseAwsResource, "month"),
	}

	c.defaultFilters = []base.Filter{
		{Key: "servicecode", Value: "AmazonEC2"},
		{Key: "productFamily", Value: "System Operation"},
		{Key: "usagetype", Value: "/EBS:VolumeP-IOPS.piops/", Operation: "REGEX"},
		{Key: "volumeApiName", Value: "gp2"},
	}

	c.valueMappings = []base.ValueMapping{
		{FromKey: "type", ToKey: "volumeApiName"},
	}

	return c
}

func (c *EbsVolumeIOPS) HourlyCost() decimal.Decimal {
	hourlyCost := c.BaseAwsPriceComponent.HourlyCost()
	iops := decimal.NewFromInt(int64(0))
	if c.AwsResource().RawValues()["iops"] != nil {
		iops = decimal.NewFromFloat(c.AwsResource().RawValues()["iops"].(float64))
	}
	return hourlyCost.Mul(iops)
}

type EbsVolume struct {
	*BaseAwsResource
}

func NewEbsVolume(address string, region string, rawValues map[string]interface{}) *EbsVolume {
	r := &EbsVolume{
		BaseAwsResource: NewBaseAwsResource(address, region, rawValues),
	}
	priceComponents := []base.PriceComponent{
		NewEbsVolumeGB("GB", r),
	}
	if r.RawValues()["type"] == "io1" {
		priceComponents = append(priceComponents, NewEbsVolumeIOPS("IOPS", r))
	}
	r.BaseAwsResource.priceComponents = priceComponents
	return r
}
