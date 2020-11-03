package schema

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/shopspring/decimal"
)

type CostComponent struct {
	Name                 string
	Unit                 string
	UnitMultiplier       int
	IgnoreIfMissingPrice bool
	ProductFilter        *ProductFilter
	PriceFilter          *PriceFilter
	HourlyQuantity       *decimal.Decimal
	MonthlyQuantity      *decimal.Decimal
	price                decimal.Decimal
	priceHash            string
	HourlyCost           *decimal.Decimal
	MonthlyCost          *decimal.Decimal
}

func (c *CostComponent) CalculateCosts() {
	c.fillQuantities()
	if c.HourlyQuantity != nil {
		c.HourlyCost = decimalPtr(c.price.Mul(*c.HourlyQuantity))
	}
	if c.MonthlyQuantity != nil {
		c.MonthlyCost = decimalPtr(c.price.Mul(*c.MonthlyQuantity))
	}
}

func (c *CostComponent) fillQuantities() {
	if c.MonthlyQuantity != nil && c.HourlyQuantity == nil {
		c.HourlyQuantity = decimalPtr(c.MonthlyQuantity.Div(hourToMonthMultiplier))
	} else if c.HourlyQuantity != nil && c.MonthlyQuantity == nil {
		c.MonthlyQuantity = decimalPtr(c.HourlyQuantity.Mul(hourToMonthMultiplier))
	}
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

func (c *CostComponent) UnitMultiplierPrice() decimal.Decimal {
	return c.Price().Mul(decimal.NewFromInt(int64(c.UnitMultiplier)))
}

func (c *CostComponent) UnitMultiplierHourlyQuantity() *decimal.Decimal {
	if c.HourlyQuantity == nil {
		return nil
	}
	m := c.HourlyQuantity.Div(decimal.NewFromInt(int64(c.UnitMultiplier)))
	return &m
}

func (c *CostComponent) UnitMultiplierMonthlyQuantity() *decimal.Decimal {
	if c.MonthlyQuantity == nil {
		return nil
	}
	m := c.MonthlyQuantity.Div(decimal.NewFromInt(int64(c.UnitMultiplier)))
	return &m
}

func (c *CostComponent) UnitWithMultiplier() string {
	s := c.Unit

	if c.UnitMultiplier != 1 {
		m := strings.ReplaceAll(humanize.SI(float64(c.UnitMultiplier), ""), " ", "")
		s = fmt.Sprintf("%s %s", m, s)
	}

	return s
}
