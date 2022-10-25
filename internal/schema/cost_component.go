package schema

import (
	"github.com/shopspring/decimal"
)

type PriceTier struct {
	Name             string
	Price            decimal.Decimal
	StartUsageAmount decimal.Decimal
	EndUsageAmount   decimal.Decimal
	HourlyQuantity   *decimal.Decimal
	MonthlyQuantity  *decimal.Decimal
	MonthlyCost      *decimal.Decimal
	HourlyCost       *decimal.Decimal
}

func (t *PriceTier) fillQuantities() {
	if t.MonthlyQuantity != nil && t.HourlyQuantity == nil {
		t.HourlyQuantity = decimalPtr(t.MonthlyQuantity.Div(HourToMonthUnitMultiplier))
	} else if t.HourlyQuantity != nil && t.MonthlyQuantity == nil {
		t.MonthlyQuantity = decimalPtr(t.HourlyQuantity.Mul(HourToMonthUnitMultiplier))
	}
}

type CostComponent struct {
	Name                  string
	Unit                  string
	UnitMultiplier        decimal.Decimal
	IgnoreIfMissingPrice  bool
	ProductFilter         *ProductFilter
	PriceFilter           *PriceFilter
	HourlyQuantity        *decimal.Decimal
	MonthlyQuantity       *decimal.Decimal
	MonthlyDiscountPerc   float64
	price                 decimal.Decimal
	priceTiers            []PriceTier
	customPrice           *decimal.Decimal
	customPriceMultiplier *decimal.Decimal
	priceHash             string
	HourlyCost            *decimal.Decimal
	MonthlyCost           *decimal.Decimal
}

func (c *CostComponent) CalculateCosts() {
	if c.priceTiers != nil {
		for i, tier := range c.priceTiers {
			if c.HourlyQuantity != nil {
				tier.HourlyQuantity = decimalPtr(decimal.NewFromInt(0))
				tier.HourlyCost = decimalPtr(decimal.NewFromInt(0))
				if tier.EndUsageAmount.GreaterThanOrEqual(*c.HourlyQuantity) && tier.StartUsageAmount.LessThan(*c.HourlyQuantity) {
					tier.HourlyQuantity = decimalPtr(c.HourlyQuantity.Sub(tier.StartUsageAmount))
				} else if tier.EndUsageAmount.GreaterThan(decimal.NewFromInt(0)) && tier.EndUsageAmount.LessThan(*c.HourlyQuantity) {
					tier.HourlyQuantity = decimalPtr(tier.EndUsageAmount.Sub(tier.StartUsageAmount))
				}
			}
			if c.MonthlyQuantity != nil {
				tier.MonthlyQuantity = decimalPtr(decimal.NewFromInt(0))
				tier.MonthlyCost = decimalPtr(decimal.NewFromInt(0))
				if tier.EndUsageAmount.GreaterThanOrEqual(*c.MonthlyQuantity) && tier.StartUsageAmount.LessThan(*c.MonthlyQuantity) {
					tier.MonthlyQuantity = decimalPtr(c.MonthlyQuantity.Sub(tier.StartUsageAmount))
				} else if tier.EndUsageAmount.GreaterThan(decimal.NewFromInt(0)) && tier.EndUsageAmount.LessThan(*c.MonthlyQuantity) {
					tier.MonthlyQuantity = decimalPtr(tier.EndUsageAmount.Sub(tier.StartUsageAmount))
				}
			}

			tier.fillQuantities()

			if tier.HourlyQuantity != nil {
				tier.HourlyCost = decimalPtr(tier.Price.Mul(*tier.HourlyQuantity))
			}
			if tier.MonthlyQuantity != nil {
				discountMul := decimal.NewFromFloat(1.0 - c.MonthlyDiscountPerc)
				tier.MonthlyCost = decimalPtr(tier.Price.Mul(*tier.MonthlyQuantity).Mul(discountMul))
			}
			c.priceTiers[i] = tier
		}
		for i, tier := range c.priceTiers {
			if c.HourlyQuantity != nil {
				if i == 0 {
					c.HourlyCost = decimalPtr(decimal.NewFromInt(0))
				}
				if tier.HourlyCost.GreaterThanOrEqual(decimal.NewFromInt(0)) {
					c.HourlyCost = decimalPtr(c.HourlyCost.Add(*tier.HourlyCost))
				}
			}
			if c.MonthlyQuantity != nil {
				if i == 0 {
					c.MonthlyCost = decimalPtr(decimal.NewFromInt(0))
				}
				if tier.MonthlyCost.GreaterThanOrEqual(decimal.NewFromInt(0)) {
					c.MonthlyCost = decimalPtr(c.MonthlyCost.Add(*tier.MonthlyCost))
				}
			}
		}
	} else {
		c.fillQuantities()
		if c.HourlyQuantity != nil {
			c.HourlyCost = decimalPtr(c.price.Mul(*c.HourlyQuantity))
		}
		if c.MonthlyQuantity != nil {
			discountMul := decimal.NewFromFloat(1.0 - c.MonthlyDiscountPerc)
			c.MonthlyCost = decimalPtr(c.price.Mul(*c.MonthlyQuantity).Mul(discountMul))
		}
	}
}

func (c *CostComponent) fillQuantities() {
	if c.MonthlyQuantity != nil && c.HourlyQuantity == nil {
		c.HourlyQuantity = decimalPtr(c.MonthlyQuantity.Div(HourToMonthUnitMultiplier))
	} else if c.HourlyQuantity != nil && c.MonthlyQuantity == nil {
		c.MonthlyQuantity = decimalPtr(c.HourlyQuantity.Mul(HourToMonthUnitMultiplier))
	}
}

func (c *CostComponent) SetPriceTiers(priceTiers []PriceTier) {
	c.priceTiers = priceTiers
}

func (c *CostComponent) PriceTiers() []PriceTier {
	return c.priceTiers
}

func (c *CostComponent) SetPrice(price decimal.Decimal) {
	c.price = price
}

func (c *CostComponent) Price() decimal.Decimal {
	return c.price
}

func (c *CostComponent) SetCustomPriceMultiplier(customPriceMultiplier *decimal.Decimal) {
	c.customPriceMultiplier = customPriceMultiplier
}

func (c *CostComponent) CustomPriceMultiplier() *decimal.Decimal {
	return c.customPriceMultiplier
}

func (c *CostComponent) SetPriceHash(priceHash string) {
	c.priceHash = priceHash
}

func (c *CostComponent) PriceHash() string {
	return c.priceHash
}

func (c *CostComponent) SetCustomPrice(price *decimal.Decimal) {
	c.customPrice = price
}

func (c *CostComponent) CustomPrice() *decimal.Decimal {
	return c.customPrice
}

func (c *CostComponent) UnitMultiplierPrice() decimal.Decimal {
	return c.Price().Mul(c.UnitMultiplier)
}

func (c *CostComponent) UnitMultiplierHourlyQuantity() *decimal.Decimal {
	if c.HourlyQuantity == nil {
		return nil
	}

	var m decimal.Decimal

	if c.UnitMultiplier.IsZero() {
		m = decimal.Zero
	} else {
		m = c.HourlyQuantity.Div(c.UnitMultiplier)
	}

	return &m
}

func (c *CostComponent) UnitMultiplierMonthlyQuantity() *decimal.Decimal {
	if c.MonthlyQuantity == nil {
		return nil
	}

	var m decimal.Decimal

	if c.UnitMultiplier.IsZero() {
		m = decimal.Zero
	} else {
		m = c.MonthlyQuantity.Div(c.UnitMultiplier)
	}

	return &m
}
