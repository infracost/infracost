package schema

import "github.com/shopspring/decimal"

// Breakdown is the state of a list of resources.
type Breakdown struct {
	Resources        []*Resource
	TotalHourlyCost  *decimal.Decimal
	TotalMonthlyCost *decimal.Decimal
}

func (b *Breakdown) calculateTotalCosts() {
	var totalHourlyCost *decimal.Decimal
	var totalMonthlyCost *decimal.Decimal

	for _, r := range b.Resources {
		if r.IsSkipped {
			continue
		}

		if r.HourlyCost != nil {
			if totalHourlyCost == nil {
				totalHourlyCost = decimalPtr(decimal.Zero)
			}

			totalHourlyCost = decimalPtr(totalHourlyCost.Add(*r.HourlyCost))
		}
		if r.MonthlyCost != nil {
			if totalMonthlyCost == nil {
				totalMonthlyCost = decimalPtr(decimal.Zero)
			}

			totalMonthlyCost = decimalPtr(totalMonthlyCost.Add(*r.MonthlyCost))
		}
	}
	b.TotalHourlyCost = totalHourlyCost
	b.TotalMonthlyCost = totalMonthlyCost
}
