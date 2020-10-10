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
	NoPrice         bool
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

func (r *Resource) RemoveCostComponent(costComponent *CostComponent) {
	n := make([]*CostComponent, 0, len(r.CostComponents)-1)
	for _, c := range r.CostComponents {
		if c != costComponent {
			n = append(n, c)
		}
	}
	r.CostComponents = n
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

func CountSkippedResources(resources []*Resource) (map[string]int, int) {
	total := 0
	typeCounts := make(map[string]int)
	for _, r := range resources {
		if r.IsSkipped && !r.NoPrice {
			total++
			if _, ok := typeCounts[r.ResourceType]; !ok {
				typeCounts[r.ResourceType] = 0
			}
			typeCounts[r.ResourceType]++
		}
	}
	return typeCounts, total
}
