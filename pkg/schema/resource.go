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
	IsSkipped      bool
	SkipMessage    string
	ResourceType   string
}

func CalculateCosts(resources []*Resource) {
	for _, r := range resources {
		r.CalculateCosts()
	}
}

func (r *Resource) CalculateCosts() {
	h := decimal.Zero

	for _, c := range r.CostComponents {
		c.CalculateCosts()
		h = h.Add(c.HourlyCost())
	}

	for _, s := range r.SubResources {
		s.CalculateCosts()
		h = h.Add(s.HourlyCost())
	}

	r.hourlyCost = h
	r.monthlyCost = h.Mul(hourToMonthMultiplier)
}

func (r *Resource) HourlyCost() decimal.Decimal {
	return r.hourlyCost
}

func (r *Resource) MonthlyCost() decimal.Decimal {
	return r.monthlyCost
}

func (r *Resource) FlattenedSubResources() []*Resource {
	resources := make([]*Resource, 0, len(r.SubResources))

	for _, s := range r.SubResources {
		resources = append(resources, s)

		if len(s.SubResources) > 0 {
			resources = append(resources, s.FlattenedSubResources()...)
		}
	}

	return resources
}

func SortResources(resources []*Resource) {
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Name < resources[j].Name
	})

	for _, r := range resources {
		SortResources(r.SubResources)

		sort.Slice(r.CostComponents, func(i, j int) bool {
			return r.CostComponents[i].Name < r.CostComponents[j].Name
		})
	}
}

func (r *Resource) IsFree() bool {
	// FIXME: Remove after https://github.com/infracost/infracost/issues/121 is done.
	return false
	return len(r.CostComponents) == 0
}
