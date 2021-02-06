package schema

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// State contains the existing, planned state of
// resources and the diff between them.
type State struct {
	ExistingState *ResourcesState
	PlannedState  *ResourcesState
	Diff          *ResourcesState
}

// AllResources returns a pointer list of all resources of the state.
func (state *State) AllResources() []*Resource {
	var resources []*Resource
	resources = append(resources, state.ExistingState.Resources...)
	resources = append(resources, state.PlannedState.Resources...)
	resources = append(resources, state.Diff.Resources...)
	return resources
}

// CalculateTotalCosts will calculate and fill the total costs fields
// of State's ResourcesState. It must be called after calculating the costs of
// the resources.
func (state *State) CalculateTotalCosts() {
	state.ExistingState.calculateTotalCosts()
	state.PlannedState.calculateTotalCosts()
}

// CalculateDiff calculates the diff of existing and planned states
// and stores it in a diff state.
func (state *State) CalculateDiff() {
	// There are many ways to calculate a diff between two sets of
	// nested objects. The method used here is to create a nested
	// hashmap of each set of states for fast lookup so the structure
	// of the hashmap would look like:
	// {resource_name.sub_resource_name: *Resource}
	// We start by traversing the existing plan. For each
	// resource that the diff is calculated, we pop it from
	// the hashmaps. In the next phase, there would be few
	// resources remaining in the planned hash map that are
	// new resources, we calculate the diff of them and the
	// diff process is over.

	existingRMap := make(map[string]*Resource, 0)
	fillResourcesMap(existingRMap, "", state.ExistingState.Resources)
	plannedRMap := make(map[string]*Resource, 0)
	fillResourcesMap(plannedRMap, "", state.PlannedState.Resources)
	state.Diff = &ResourcesState{}

	for _, resource := range state.ExistingState.Resources {
		resourceKey := resource.Name
		changed, diff := diffResourcesByKey(resourceKey, existingRMap, plannedRMap)
		if changed {
			state.Diff.Resources = append(state.Diff.Resources, diff)
		}
	}

	for _, resource := range state.PlannedState.Resources {
		resourceKey := resource.Name
		if _, ok := plannedRMap[resourceKey]; !ok {
			continue
		}
		changed, diff := diffResourcesByKey(resourceKey, existingRMap, plannedRMap)
		if changed {
			state.Diff.Resources = append(state.Diff.Resources, diff)
		}
	}

}

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
	//
	ccChanged, ccDiff := diffCostComponentsByResource(existing, planned)
	if ccChanged {
		diff.CostComponents = ccDiff
		changed = true
	}
	//
	if existingOk {
		delete(existingResMap, resourceKey)
	}
	if plannedOk {
		delete(plannedResMap, resourceKey)
	}

	return changed, diff
}

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

func getCostComponentsMap(resource *Resource) map[string]*CostComponent {
	result := make(map[string]*CostComponent)
	for _, costComponent := range resource.CostComponents {
		result[costComponent.Name] = costComponent
	}
	return result
}
