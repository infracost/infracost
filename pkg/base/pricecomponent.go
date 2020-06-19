package base

import (
	"github.com/shopspring/decimal"
)

type PriceComponent interface {
	Name() string
	ShouldSkip() bool
	GetFilters() []Filter
	CalculateHourlyCost(decimal.Decimal) decimal.Decimal
}

type BasePriceComponent struct {
	name         string
	resource     Resource
	priceMapping *PriceMapping
}

func NewBasePriceComponent(name string, resource Resource, priceMapping *PriceMapping) *BasePriceComponent {
	c := &BasePriceComponent{
		name:         name,
		resource:     resource,
		priceMapping: priceMapping,
	}

	return c
}

func (c *BasePriceComponent) Name() string {
	return c.name
}

func (c *BasePriceComponent) ShouldSkip() bool {
	if c.priceMapping.ShouldSkip != nil {
		return c.priceMapping.ShouldSkip(c.resource.RawValues())
	}
	return false
}

func (c *BasePriceComponent) GetFilters() []Filter {
	return MergeFilters(c.resource.GetFilters(), c.priceMapping.GetFilters(c.resource.RawValues()))
}

func (c *BasePriceComponent) CalculateHourlyCost(price decimal.Decimal) decimal.Decimal {
	var cost decimal.Decimal
	if c.priceMapping.CalculateCost != nil {
		cost = c.priceMapping.CalculateCost(price)
	} else {
		cost = price
	}

	timeUnitSecs := map[string]int64{
		"hour":  60 * 60,
		"month": 60 * 60 * 730,
	}
	timeUnitMultiplier := timeUnitSecs["hour"] / timeUnitSecs[c.priceMapping.TimeUnit]

	hourlyCost := cost.Mul(decimal.NewFromInt(timeUnitMultiplier))
	return hourlyCost
}
