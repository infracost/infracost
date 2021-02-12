package schema

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// Project contains the existing, planned state of
// resources and the diff between them.
type Project struct {
	PastBreakdown *Breakdown
	Breakdown     *Breakdown
	Diff          *Breakdown
}

func NewProject() *Project {
	return &Project{
		PastBreakdown: &Breakdown{},
		Breakdown:     &Breakdown{},
		Diff:          &Breakdown{},
	}
}

// AllResources returns a pointer list of all resources of the state.
func (state *Project) AllResources() []*Resource {
	var resources []*Resource
	resources = append(resources, state.PastBreakdown.Resources...)
	resources = append(resources, state.Breakdown.Resources...)
	resources = append(resources, state.Diff.Resources...)
	return resources
}

// CalculateTotalCosts will calculate and fill the total costs fields
// of Project's Breakdown. It must be called after calculating the costs of
// the resources.
func (state *Project) CalculateTotalCosts() {
	state.PastBreakdown.calculateTotalCosts()
	state.Breakdown.calculateTotalCosts()
}

// CalculateDiff calculates the diff of existing and planned states
// and stores it in a diff state.
func (state *Project) CalculateDiff() {
	// There are many ways to calculate a diff between two sets of
	// nested objects. The method used here is to create a nested
	// hashmap of each set of states for fast lookup so the structure
	// of the hashmap would look like:
	// {resource_name.sub_resource_name: *Resource}
	// We start by traversing the existing plan. For each
	// resource that the diff is calculated, we pop it from
	// the hashmaps. In the next phase, there would be few
	// resources remaining in the planned hash map that are
	// new resources, we then traverse the planned resources
	// and check their existence in the planned hash map. If they
	// are found it means they are new resources and we need to
	// calculate the diff for them. This way a complete diff for
	// all resources is calculated.

	existingRMap := make(map[string]*Resource)
	fillResourcesMap(existingRMap, "", state.PastBreakdown.Resources)
	plannedRMap := make(map[string]*Resource)
	fillResourcesMap(plannedRMap, "", state.Breakdown.Resources)
	state.Diff = &Breakdown{
		Resources:        []*Resource{},
		TotalHourlyCost:  &decimal.Zero,
		TotalMonthlyCost: &decimal.Zero,
	}

	for _, resource := range state.PastBreakdown.Resources {
		resourceKey := resource.Name
		changed, diff := diffResourcesByKey(resourceKey, existingRMap, plannedRMap)
		if changed {
			state.Diff.Resources = append(state.Diff.Resources, diff)
		}
	}

	for _, resource := range state.Breakdown.Resources {
		resourceKey := resource.Name
		if _, ok := plannedRMap[resourceKey]; !ok {
			continue
		}
		changed, diff := diffResourcesByKey(resourceKey, existingRMap, plannedRMap)
		if changed {
			state.Diff.Resources = append(state.Diff.Resources, diff)
		}
	}

	state.Diff.TotalHourlyCost = diffDecimals(state.Breakdown.TotalHourlyCost, state.PastBreakdown.TotalHourlyCost)
	state.Diff.TotalMonthlyCost = diffDecimals(state.Breakdown.TotalMonthlyCost, state.PastBreakdown.TotalMonthlyCost)

}

// diffResourcesByKey calculates the diff between two resources given their resourcesMap and
// their key.
func diffResourcesByKey(resourceKey string, existingResMap, plannedResMap map[string]*Resource) (bool, *Resource) {
	existing, existingOk := existingResMap[resourceKey]
	planned, plannedOk := plannedResMap[resourceKey]
	baseResource := planned
	if planned == nil {
		baseResource = existing
		planned = &Resource{}
	}
	if existing == nil {
		existing = &Resource{}
	}
	changed := false
	diff := &Resource{
		Name:         baseResource.Name,
		IsSkipped:    baseResource.IsSkipped,
		NoPrice:      baseResource.NoPrice,
		SkipMessage:  baseResource.SkipMessage,
		ResourceType: baseResource.ResourceType,
		Tags:         baseResource.Tags,

		HourlyCost:  diffDecimals(planned.HourlyCost, existing.HourlyCost),
		MonthlyCost: diffDecimals(planned.MonthlyCost, existing.MonthlyCost),
	}
	for _, subResource := range existing.SubResources {
		subKey := fmt.Sprintf("%v.%v", resourceKey, subResource.Name)
		subChanged, subDiff := diffResourcesByKey(subKey, existingResMap, plannedResMap)
		if subChanged {
			diff.SubResources = append(diff.SubResources, subDiff)
			changed = true
		}
	}
	for _, subResource := range planned.SubResources {
		subKey := fmt.Sprintf("%v.%v", resourceKey, subResource.Name)
		if _, ok := plannedResMap[subKey]; !ok {
			continue
		}
		subChanged, subDiff := diffResourcesByKey(subKey, existingResMap, plannedResMap)
		if subChanged {
			diff.SubResources = append(diff.SubResources, subDiff)
			changed = true
		}
	}
	ccChanged, ccDiff := diffCostComponentsByResource(existing, planned)
	if ccChanged {
		diff.CostComponents = ccDiff
		changed = true
	}
	if existingOk {
		delete(existingResMap, resourceKey)
	}
	if plannedOk {
		delete(plannedResMap, resourceKey)
	}

	return changed, diff
}

// diffCostComponentsByResource calculates the diff of cost components of two resource.
// It uses the same strategy as the calculating the diff of resources in the CalculateDiff func.
func diffCostComponentsByResource(existing, planned *Resource) (bool, []*CostComponent) {
	result := make([]*CostComponent, 0)
	changed := false
	existingCCMap := getCostComponentsMap(existing)
	plannedCCMap := getCostComponentsMap(planned)
	for _, costComponent := range existing.CostComponents {
		key := costComponent.Name
		changed, diff := diffCostComponentsByKey(key, existingCCMap, plannedCCMap)
		if changed {
			result = append(result, diff)
		}
	}
	for _, costComponent := range planned.CostComponents {
		key := costComponent.Name
		if _, ok := plannedCCMap[key]; !ok {
			continue
		}
		changed, diff := diffCostComponentsByKey(key, existingCCMap, plannedCCMap)
		if changed {
			result = append(result, diff)
		}
	}
	if len(result) > 0 {
		changed = true
	}
	return changed, result
}

// diffCostComponentsByKey calculates the diff between two cost components given
// their costComponentsMap and their key.
func diffCostComponentsByKey(key string, existingCCMap, plannedCCMap map[string]*CostComponent) (bool, *CostComponent) {
	existing, existingOk := existingCCMap[key]
	planned, plannedOk := plannedCCMap[key]
	if planned == nil && existing == nil {
		return false, nil
	}
	baseCostComponent := planned
	if planned == nil {
		baseCostComponent = existing
		planned = &CostComponent{}
	}
	if existing == nil {
		existing = &CostComponent{}
	}
	changed := false
	diff := &CostComponent{
		Name:                 baseCostComponent.Name,
		Unit:                 baseCostComponent.Unit,
		UnitMultiplier:       baseCostComponent.UnitMultiplier,
		IgnoreIfMissingPrice: baseCostComponent.IgnoreIfMissingPrice,
		ProductFilter:        baseCostComponent.ProductFilter,
		PriceFilter:          baseCostComponent.PriceFilter,
		priceHash:            baseCostComponent.priceHash,

		HourlyQuantity:      diffDecimals(planned.HourlyQuantity, existing.HourlyQuantity),
		MonthlyQuantity:     diffDecimals(planned.MonthlyQuantity, existing.MonthlyQuantity),
		MonthlyDiscountPerc: planned.MonthlyDiscountPerc - existing.MonthlyDiscountPerc,
		price:               *diffDecimals(&planned.price, &existing.price),
		HourlyCost:          diffDecimals(planned.HourlyCost, existing.HourlyCost),
		MonthlyCost:         diffDecimals(planned.MonthlyCost, existing.MonthlyCost),
	}
	if !diff.HourlyQuantity.IsZero() || !diff.MonthlyQuantity.IsZero() ||
		diff.MonthlyDiscountPerc != 0 || !diff.price.IsZero() ||
		!diff.HourlyCost.IsZero() || !diff.MonthlyCost.IsZero() {
		changed = true
	}
	if existingOk {
		delete(existingCCMap, key)
	}
	if plannedOk {
		delete(plannedCCMap, key)
	}

	return changed, diff
}

// diffDecimals calculates the diff between two decimals.
func diffDecimals(planned *decimal.Decimal, existing *decimal.Decimal) *decimal.Decimal {
	var diff decimal.Decimal
	if existing == nil && planned == nil {
		diff = decimal.Zero
	} else if existing == nil {
		diff = *planned
	} else if planned == nil {
		diff = existing.Neg()
	} else if planned.Equals(*existing) {
		// Handling the strange behavior of decimal for easier testing.
		diff = decimal.Zero
	} else {
		diff = planned.Sub(*existing)
	}
	return &diff
}

// fillResourcesMap fills a given resource map with the structure: {resource_name.sub_resource_name: *Resource}
func fillResourcesMap(resourcesMap map[string]*Resource, rootKey string, resources []*Resource) {
	for _, resource := range resources {
		key := resource.Name
		if rootKey != "" {
			key = fmt.Sprintf("%v.%v", rootKey, resource.Name)
		}
		resourcesMap[key] = resource
		fillResourcesMap(resourcesMap, key, resource.SubResources)
	}
}

// fillResourcesMap creates a cost components map with the structure: {cost_component_name: *CostComponent}
func getCostComponentsMap(resource *Resource) map[string]*CostComponent {
	result := make(map[string]*CostComponent)
	for _, costComponent := range resource.CostComponents {
		result[costComponent.Name] = costComponent
	}
	return result
}

func AllResources(projects []*Project) []*Resource {
	resources := make([]*Resource, 0)

	for _, p := range projects {
		resources = append(resources, p.Breakdown.Resources...)
	}

	return resources
}
