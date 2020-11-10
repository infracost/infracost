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
	HourlyCost     *decimal.Decimal
	MonthlyCost    *decimal.Decimal
	IsSkipped      bool
	NoPrice        bool
	SkipMessage    string
	ResourceType   string
}

type ResourceSummary struct {
	SupportedCounts   map[string]int `json:"supportedCounts"`
	UnsupportedCounts map[string]int `json:"unsupportedCounts"`
	TotalSupported    int            `json:"totalSupported"`
	TotalUnsupported  int            `json:"totalUnsupported"`
	TotalNoPrice      int            `json:"totalNoPrice"`
	Total             int            `json:"total"`
}

func CalculateCosts(resources []*Resource) {
	for _, r := range resources {
		r.CalculateCosts()
	}
}

func (r *Resource) CalculateCosts() {
	h := decimal.Zero
	m := decimal.Zero
	hasCost := false

	for _, c := range r.CostComponents {
		c.CalculateCosts()
		if c.HourlyCost != nil || c.MonthlyCost != nil {
			hasCost = true
		}
		if c.HourlyCost != nil {
			h = h.Add(*c.HourlyCost)
		}
		if c.MonthlyCost != nil {
			m = m.Add(*c.MonthlyCost)
		}
	}

	for _, s := range r.SubResources {
		s.CalculateCosts()
		if s.HourlyCost != nil || s.MonthlyCost != nil {
			hasCost = true
		}
		if s.HourlyCost != nil {
			h = h.Add(*s.HourlyCost)
		}
		if s.MonthlyCost != nil {
			m = m.Add(*s.MonthlyCost)
		}
	}

	if hasCost {
		r.HourlyCost = &h
		r.MonthlyCost = &m
	}
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
}

func GenerateResourceSummary(resources []*Resource) *ResourceSummary {
	supportedCounts := make(map[string]int)
	unsupportedCounts := make(map[string]int)
	totalSupported := 0
	totalUnsupported := 0
	totalNoPrice := 0

	for _, r := range resources {
		if r.NoPrice {
			totalNoPrice++
		} else if r.IsSkipped {
			totalUnsupported++
			if _, ok := unsupportedCounts[r.ResourceType]; !ok {
				unsupportedCounts[r.ResourceType] = 0
			}
			unsupportedCounts[r.ResourceType]++
		} else {
			totalSupported++
			if _, ok := supportedCounts[r.ResourceType]; !ok {
				supportedCounts[r.ResourceType] = 0
			}
			supportedCounts[r.ResourceType]++
		}
	}

	return &ResourceSummary{
		SupportedCounts:   supportedCounts,
		UnsupportedCounts: unsupportedCounts,
		TotalSupported:    totalSupported,
		TotalUnsupported:  totalUnsupported,
		TotalNoPrice:      totalNoPrice,
		Total:             len(resources),
	}
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}
