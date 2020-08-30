package costs

import (
	"infracost/pkg/schema"

	"github.com/shopspring/decimal"
)

var hourToMonthMultiplier = decimal.NewFromInt(730)

type CostComponent struct {
	schemaCostComponent *schema.CostComponent
	hourlyCost          decimal.Decimal
	monthlyCost         decimal.Decimal
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
	return c.hourlyCost
}

func (c *CostComponent) MonthlyCost() decimal.Decimal {
	return c.monthlyCost
}
