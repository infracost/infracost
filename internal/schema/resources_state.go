package schema

import "github.com/shopspring/decimal"

// ResourcesState is the state of a list of resources.
type ResourcesState struct {
	Resources        []*Resource
	TotalHourlyCost  *decimal.Decimal
	TotalMonthlyCost *decimal.Decimal
}

func (rs *ResourcesState) calculateTotalCosts() {
	var totalHourlyCost *decimal.Decimal
	var totalMonthlyCost *decimal.Decimal

	for _, r := range rs.Resources {
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
	rs.TotalHourlyCost = totalHourlyCost
	rs.TotalMonthlyCost = totalMonthlyCost
}
