package schema

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// CalculateDiff calculates the diff of past and current resources
func calculateDiff(past []*Resource, current []*Resource) []*Resource {
	// There are many ways to calculate a diff between two sets of
	// nested objects. The method used here is to create a nested
	// hashmap of each set of states for fast lookup so the structure
	// of the hashmap would look like:
	// {resource_name.sub_resource_name: *Resource}
	// We start by traversing the past plan. For each
	// resource that the diff is calculated, we pop it from
	// the hashmaps. In the next phase, there would be few
	// resources remaining in the current hash map that are
	// new resources, we then traverse the current resources
	// and check their existence in the current hash map. If they
	// are found it means they are new resources and we need to
	// calculate the diff for them. This way a complete diff for
	// all resources is calculated.

	pastRMap := make(map[string]*Resource)
	fillResourcesMap(pastRMap, "", past)
	currentRMap := make(map[string]*Resource)
	fillResourcesMap(currentRMap, "", current)

	diff := make([]*Resource, 0)

	for _, resource := range past {
		resourceKey := resource.Name
		changed, resources := diffResourcesByKey(resourceKey, pastRMap, currentRMap)
		if changed {
			diff = append(diff, resources)
		}
	}

	for _, resource := range current {
		resourceKey := resource.Name
		if _, ok := currentRMap[resourceKey]; !ok {
			continue
		}
		changed, resources := diffResourcesByKey(resourceKey, pastRMap, currentRMap)
		if changed {
			diff = append(diff, resources)
		}
	}

	return diff
}

// diffResourcesByKey calculates the diff between two resources given their resourcesMap and
// their key.
func diffResourcesByKey(resourceKey string, pastResMap, currentResMap map[string]*Resource) (bool, *Resource) {
	past, pastOk := pastResMap[resourceKey]
	current, currentOk := currentResMap[resourceKey]
	if current == nil && past == nil {
		return false, nil
	}
	baseResource := current
	if current == nil {
		baseResource = past
		current = &Resource{}
	}
	if past == nil {
		past = &Resource{}
	}
	changed := false
	diff := &Resource{
		Name:         baseResource.Name,
		IsSkipped:    baseResource.IsSkipped,
		NoPrice:      baseResource.NoPrice,
		SkipMessage:  baseResource.SkipMessage,
		ResourceType: baseResource.ResourceType,
		Tags:         baseResource.Tags,

		HourlyCost:  diffDecimals(current.HourlyCost, past.HourlyCost),
		MonthlyCost: diffDecimals(current.MonthlyCost, past.MonthlyCost),
	}
	for _, subResource := range past.SubResources {
		subKey := fmt.Sprintf("%v.%v", resourceKey, subResource.Name)
		subChanged, subDiff := diffResourcesByKey(subKey, pastResMap, currentResMap)
		if subChanged {
			diff.SubResources = append(diff.SubResources, subDiff)
			changed = true
		}
	}
	for _, subResource := range current.SubResources {
		subKey := fmt.Sprintf("%v.%v", resourceKey, subResource.Name)
		if _, ok := currentResMap[subKey]; !ok {
			continue
		}
		subChanged, subDiff := diffResourcesByKey(subKey, pastResMap, currentResMap)
		if subChanged {
			diff.SubResources = append(diff.SubResources, subDiff)
			changed = true
		}
	}
	ccChanged, ccDiff := diffCostComponentsByResource(past, current)
	if ccChanged {
		diff.CostComponents = ccDiff
		changed = true
	}
	if pastOk {
		delete(pastResMap, resourceKey)
	}
	if currentOk {
		delete(currentResMap, resourceKey)
	}

	return changed, diff
}

// diffCostComponentsByResource calculates the diff of cost components of two resource.
// It uses the same strategy as the calculating the diff of resources in the CalculateDiff func.
func diffCostComponentsByResource(past, current *Resource) (bool, []*CostComponent) {
	result := make([]*CostComponent, 0)
	changed := false
	pastCCMap := getCostComponentsMap(past)
	currentCCMap := getCostComponentsMap(current)
	for _, costComponent := range past.CostComponents {
		key := costComponent.Name
		changed, diff := diffCostComponentsByKey(key, pastCCMap, currentCCMap)
		if changed {
			result = append(result, diff)
		}
	}
	for _, costComponent := range current.CostComponents {
		key := costComponent.Name
		if _, ok := currentCCMap[key]; !ok {
			continue
		}
		changed, diff := diffCostComponentsByKey(key, pastCCMap, currentCCMap)
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
func diffCostComponentsByKey(key string, pastCCMap, currentCCMap map[string]*CostComponent) (bool, *CostComponent) {
	past, pastOk := pastCCMap[key]
	current, currentOk := currentCCMap[key]
	if current == nil && past == nil {
		return false, nil
	}
	baseCostComponent := current
	if current == nil {
		baseCostComponent = past
		current = &CostComponent{}
	}
	if past == nil {
		past = &CostComponent{}
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

		HourlyQuantity:      diffDecimals(current.HourlyQuantity, past.HourlyQuantity),
		MonthlyQuantity:     diffDecimals(current.MonthlyQuantity, past.MonthlyQuantity),
		MonthlyDiscountPerc: current.MonthlyDiscountPerc - past.MonthlyDiscountPerc,
		price:               *diffDecimals(&current.price, &past.price),
		HourlyCost:          diffDecimals(current.HourlyCost, past.HourlyCost),
		MonthlyCost:         diffDecimals(current.MonthlyCost, past.MonthlyCost),
	}
	if !diff.HourlyQuantity.IsZero() || !diff.MonthlyQuantity.IsZero() ||
		diff.MonthlyDiscountPerc != 0 || !diff.price.IsZero() ||
		!diff.HourlyCost.IsZero() || !diff.MonthlyCost.IsZero() {
		changed = true
	}
	if pastOk {
		delete(pastCCMap, key)
	}
	if currentOk {
		delete(currentCCMap, key)
	}

	return changed, diff
}

// diffDecimals calculates the diff between two decimals.
func diffDecimals(current *decimal.Decimal, past *decimal.Decimal) *decimal.Decimal {
	var diff decimal.Decimal
	if past == nil && current == nil {
		diff = decimal.Zero
	} else if past == nil {
		diff = *current
	} else if current == nil {
		diff = past.Neg()
	} else if current.Equals(*past) {
		// Handling the strange behavior of decimal for easier testing.
		diff = decimal.Zero
	} else {
		diff = current.Sub(*past)
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
