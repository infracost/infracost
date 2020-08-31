package schema

import (
	"github.com/shopspring/decimal"
)

type CostComponent struct {
	Name            string
	Unit            string
	ProductFilter   *ProductFilter
	PriceFilter     *PriceFilter
	HourlyQuantity  *decimal.Decimal
	MonthlyQuantity *decimal.Decimal
	price           decimal.Decimal
	priceHash       string
	hourlyCost      decimal.Decimal
	monthlyCost     decimal.Decimal
}

func (c *CostComponent) CalculateCosts() {
	c.fillQuantities()
	c.hourlyCost = c.price.Mul(*c.HourlyQuantity)
	c.monthlyCost = c.price.Mul(*c.MonthlyQuantity)
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

func (c *CostComponent) fillQuantities() {
	if c.HourlyQuantity == nil && c.MonthlyQuantity == nil {
		c.HourlyQuantity = decimalPtr(decimal.Zero)
		c.MonthlyQuantity = decimalPtr(decimal.Zero)
	} else if c.HourlyQuantity == nil {
		c.HourlyQuantity = decimalPtr(c.MonthlyQuantity.Div(hourToMonthMultiplier))
	} else if c.MonthlyQuantity == nil {
		c.MonthlyQuantity = decimalPtr(c.HourlyQuantity.Mul(hourToMonthMultiplier))
	}
}

func (c *CostComponent) HourlyCost() decimal.Decimal {
	return c.hourlyCost
}

func (c *CostComponent) MonthlyCost() decimal.Decimal {
	return c.monthlyCost
}

func (c *CostComponent) SetPrice(price decimal.Decimal) {
	c.price = price
}

func (c *CostComponent) Price() decimal.Decimal {
	return c.price
}

func (c *CostComponent) SetPriceHash(priceHash string) {
	c.priceHash = priceHash
}

func (c *CostComponent) PriceHash() string {
	return c.priceHash
}
