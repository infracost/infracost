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

func CalculateCosts(r []*Resource) {
	for _, res := range r {
		res.CalculateCosts()
	}
}

func (r *Resource) CalculateCosts() {
	h := decimal.Zero

	for _, c := range r.CostComponents {
		c.CalculateCosts()
		h = h.Add(c.HourlyCost())
	}

	for _, r := range r.SubResources {
		r.CalculateCosts()
		h = h.Add(r.HourlyCost())
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
	res := make([]*Resource, 0, len(r.SubResources))

	for _, s := range r.SubResources {
		res = append(res, s)

		if len(s.SubResources) > 0 {
			res = append(res, s.FlattenedSubResources()...)
		}
	}

	return res
}

func SortResources(r []*Resource) {
	sort.Slice(r, func(i, j int) bool {
		return r[i].Name < r[j].Name
	})

	for _, res := range r {
		SortResources(res.SubResources)

		sort.Slice(res.CostComponents, func(i, j int) bool {
			return res.CostComponents[i].Name < res.CostComponents[j].Name
		})
	}
}
