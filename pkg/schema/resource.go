package schema

import (
	"sort"

	"github.com/shopspring/decimal"
)

var hourToMonthMultiplier = decimal.NewFromInt(730)

type ResourceFunc func(*ResourceData, *ResourceData) *Resource

type Resource struct {
	Name           string
	CostComponents []*CostComponent
	SubResources   []*Resource
	hourlyCost     decimal.Decimal
	monthlyCost    decimal.Decimal
}

func CalculateCosts(resources []*Resource) {
	for _, resource := range resources {
		resource.CalculateCosts()
	}
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
	res := make([]*Resource, 0, len(r.SubResources))

	for _, s := range r.SubResources {
		res = append(res, s)

		if len(s.SubResources) > 0 {
			res = append(res, s.FlattenedSubResources()...)
		}
	}

	return res
}

func SortResources(resources []*Resource) {
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Name < resources[j].Name
	})

	for _, resource := range resources {
		SortResources(resource.SubResources)

		sort.Slice(resource.CostComponents, func(i, j int) bool {
			return resource.CostComponents[i].Name < resource.CostComponents[j].Name
		})
	}
}
