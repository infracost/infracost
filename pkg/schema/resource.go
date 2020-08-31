package schema

import (
	"github.com/shopspring/decimal"
)

var hourToMonthMultiplier = decimal.NewFromInt(730)

type ResourceFunc func(*ResourceData) *Resource

type Resource struct {
	Name           string
	SubResources   []*Resource
	CostComponents []*CostComponent
	hourlyCost     decimal.Decimal
	monthlyCost    decimal.Decimal
}

func (r *Resource) CalculateCosts() {
	hourlyCost := decimal.Zero

	for _, costComponent := range r.CostComponents {
		costComponent.CalculateCosts()
		hourlyCost = hourlyCost.Add(costComponent.HourlyCost())
	}

	for _, subResource := range r.SubResources {
		subResource.CalculateCosts()
		hourlyCost = hourlyCost.Add(subResource.HourlyCost())
	}

	r.hourlyCost = hourlyCost
	r.monthlyCost = hourlyCost.Mul(hourToMonthMultiplier)
}

func (r *Resource) HourlyCost() decimal.Decimal {
	return r.hourlyCost
}

func (r *Resource) MonthlyCost() decimal.Decimal {
	return r.monthlyCost
}

func (r *Resource) FlattenedSubResources() []*Resource {
	subResources := make([]*Resource, 0, len(r.SubResources))
	for _, subResource := range r.SubResources {
		subResources = append(subResources, subResource)
		if len(subResource.SubResources) > 0 {
			subResources = append(subResources, subResource.FlattenedSubResources()...)
		}
	}
	return subResources
}
