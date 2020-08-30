package costs

import (
	"infracost/pkg/schema"

	"github.com/shopspring/decimal"
)

type CostComponent struct {
	schemaCostComponent *schema.CostComponent
	price               decimal.Decimal
	priceHash           string
}

func NewCostComponent(schemaCostComponent *schema.CostComponent) *CostComponent {
	return &CostComponent{
		schemaCostComponent: schemaCostComponent,
	}
}

func (c *CostComponent) Name() string {
	return c.schemaCostComponent.Name
}

func (c *CostComponent) Unit() string {
	return c.schemaCostComponent.Unit
}

func (c *CostComponent) HourlyQuantity() decimal.Decimal {
	if c.schemaCostComponent.HourlyQuantity == nil {
		if c.schemaCostComponent.MonthlyQuantity == nil {
			return decimal.Zero
		}
		return c.schemaCostComponent.MonthlyQuantity.Div(hourToMonthMultiplier)
	}
	return *c.schemaCostComponent.HourlyQuantity
}

func (c *CostComponent) MonthlyQuantity() decimal.Decimal {
	if c.schemaCostComponent.MonthlyQuantity == nil {
		if c.schemaCostComponent.HourlyQuantity == nil {
			return decimal.Zero
		}
		return c.schemaCostComponent.HourlyQuantity.Mul(hourToMonthMultiplier)
	}
	return *c.schemaCostComponent.MonthlyQuantity
}

func (c *CostComponent) HourlyCost() decimal.Decimal {
	return c.price.Mul(c.HourlyQuantity())
}

func (c *CostComponent) MonthlyCost() decimal.Decimal {
	return c.price.Mul(c.MonthlyQuantity())
}

func (c *CostComponent) SetPrice(price decimal.Decimal) {
	c.price = price
}

func (c *CostComponent) SetPriceHash(priceHash string) {
	c.priceHash = priceHash
}
