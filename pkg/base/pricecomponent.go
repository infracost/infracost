package base

import (
	"github.com/shopspring/decimal"
)

type PriceComponent interface {
	Name() string
	Resource() Resource
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

func (c *BasePriceComponent) Resource() Resource {
	return c.resource
}

func (c *BasePriceComponent) ShouldSkip() bool {
	if c.priceMapping.ShouldSkip != nil {
		return c.priceMapping.ShouldSkip(c.resource.RawValues())
	}
	return false
}

func (c *BasePriceComponent) GetFilters() []Filter {
	return MergeFilters(c.resource.GetFilters(), c.priceMapping.GetFilters(c.resource))
}

func (c *BasePriceComponent) CalculateHourlyCost(price decimal.Decimal) decimal.Decimal {
	var cost decimal.Decimal
	if c.priceMapping.CalculateCost != nil {
		cost = c.priceMapping.CalculateCost(price, c.resource)
	} else {
		cost = price
	}

	timeUnitSecs := map[string]decimal.Decimal{
		"hour":  decimal.NewFromInt(int64(60 * 60)),
		"month": decimal.NewFromInt(int64(60 * 60 * 730)),
	}
	timeUnitMultiplier := timeUnitSecs["hour"].Div(timeUnitSecs[c.priceMapping.TimeUnit])

	hourlyCost := cost.Mul(timeUnitMultiplier)

	hourlyCost = c.Resource().AdjustCost(hourlyCost)

	return hourlyCost
}
