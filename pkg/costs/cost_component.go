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

func (r *CostComponent) Name() string {
	return r.schemaCostComponent.Name
}

func (r *CostComponent) Unit() string {
	return r.schemaCostComponent.Unit
}

func (r *CostComponent) HourlyQuantity() decimal.Decimal {
	if r.schemaCostComponent.HourlyQuantity == nil {
		if r.schemaCostComponent.MonthlyQuantity == nil {
			return decimal.Zero
		}
		return r.schemaCostComponent.MonthlyQuantity.Div(hourToMonthMultiplier)
	}
	return *r.schemaCostComponent.HourlyQuantity
}

func (r *CostComponent) MonthlyQuantity() decimal.Decimal {
	if r.schemaCostComponent.MonthlyQuantity == nil {
		if r.schemaCostComponent.HourlyQuantity == nil {
			return decimal.Zero
		}
		return r.schemaCostComponent.HourlyQuantity.Mul(hourToMonthMultiplier)
	}
	return *r.schemaCostComponent.MonthlyQuantity
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
