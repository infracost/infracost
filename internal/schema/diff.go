package schema

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/logging"
)

// nameBracketReg matches the part of a cost component name before the brackets, and the part in the brackets
var nameBracketReg = regexp.MustCompile(`(.*?)\s*\((.*?)\)`)

// CalculateDiff calculates the diff of past and current resources
func CalculateDiff(past []*Resource, current []*Resource) []*Resource {
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
		logging.Logger.Debug().Msgf("diffResourcesByKey nil current and past with key %s", resourceKey)
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
		DefaultTags:  baseResource.DefaultTags,

		HourlyCost:       diffDecimals(current.HourlyCost, past.HourlyCost),
		MonthlyCost:      diffDecimals(current.MonthlyCost, past.MonthlyCost),
		MonthlyUsageCost: diffDecimals(current.MonthlyUsageCost, past.MonthlyUsageCost),
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
func diffCostComponentsByResource(pastResource, currentResource *Resource) (bool, []*CostComponent) {
	result := make([]*CostComponent, 0)
	changed := false

	remainingCostComponents := map[string]*CostComponent{}
	for _, costComponent := range currentResource.CostComponents {
		remainingCostComponents[costComponent.Name] = costComponent
	}

	for _, pastCostComponent := range pastResource.CostComponents {
		var currentCostComponent *CostComponent
		if currentResource != nil {
			currentCostComponent = findMatchingCostComponent(currentResource.CostComponents, pastCostComponent.Name)
		}
		if currentCostComponent != nil {
			delete(remainingCostComponents, currentCostComponent.Name)
		}

		changed, diff := diffCostComponents(pastCostComponent, currentCostComponent)
		if changed {
			result = append(result, diff)
		}
	}

	// Loop through all the current ones so we maintain the order
	for _, currentCostComponent := range currentResource.CostComponents {
		// Skip any that have already been processed
		if _, ok := remainingCostComponents[currentCostComponent.Name]; !ok {
			continue
		}

		// Since we've skipped the processed ones then there will be no past cost component to diff against
		changed, diff := diffCostComponents(nil, currentCostComponent)
		if changed {
			result = append(result, diff)
		}
	}
	if len(result) > 0 {
		changed = true
	}
	return changed, result
}

// diffCostComponents creates a new cost component which contains the diff of the two cost components
func diffCostComponents(past *CostComponent, current *CostComponent) (bool, *CostComponent) {
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
		Name:                 diffName(current.Name, past.Name),
		Unit:                 baseCostComponent.Unit,
		UnitMultiplier:       baseCostComponent.UnitMultiplier,
		IgnoreIfMissingPrice: baseCostComponent.IgnoreIfMissingPrice,
		ProductFilter:        baseCostComponent.ProductFilter,
		PriceFilter:          baseCostComponent.PriceFilter,
		priceHash:            baseCostComponent.priceHash,
		UsageBased:           baseCostComponent.UsageBased,

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

	return changed, diff
}

// findMatchingCostComponent finds a matching cost component by first looking for an exact match by name
// and if that's not found, looking for a match of everything before any brackets.
func findMatchingCostComponent(costComponents []*CostComponent, name string) *CostComponent {
	for _, costComponent := range costComponents {
		if costComponent.Name == name {
			return costComponent
		}
	}

	for _, costComponent := range costComponents {
		splitKey := strings.Split(name, " (")
		splitName := strings.Split(costComponent.Name, " (")
		if len(splitKey) > 1 && len(splitName) > 1 && splitName[0] == splitKey[0] {
			return costComponent
		}
	}

	return nil
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

// diffName creates a new cost component name for the diff cost component based on the existing cost components.
// Anything that is in brackets is treated as a label and any difference in the labels across the past and current
// are represented as "old → new"
//
// For example:
// If past is "Instance usage (Linux/UNIX, on-demand, t3.small)"
// and current is "Instance usage (Linux/UNIX, on-demand, t3.medium)"
// this returns "Instance usage (Linux/UNIX, on-demand, t3.small → t3.medium)"
func diffName(current string, past string) string {
	if current == "" {
		return past
	}

	if past == "" {
		return current
	}

	currentM := nameBracketReg.FindStringSubmatch(current)
	pastM := nameBracketReg.FindStringSubmatch(past)

	if len(currentM) < 3 || len(pastM) < 3 {
		return current
	}

	pastLabels := strings.Split(pastM[2], ", ")
	currentLabels := strings.Split(currentM[2], ", ")

	// If the names don't have the same label count then return the labels in the format `(old, labels) → (new, labels)`
	if len(pastLabels) != len(currentLabels) {
		return fmt.Sprintf("%s (%s) → (%s)", currentM[1], pastM[2], currentM[2])
	}

	labelCount := len(currentLabels)
	labels := make([]string, 0, labelCount)

	for i := range labelCount {
		if i > len(pastLabels)-1 {
			labels = append(labels, currentLabels[i])
		} else if i > len(currentLabels)-1 {
			labels = append(labels, pastLabels[i])
		} else if pastLabels[i] == currentLabels[i] {
			labels = append(labels, currentLabels[i])
		} else {
			labels = append(labels, fmt.Sprintf("%s → %s", pastLabels[i], currentLabels[i]))
		}
	}

	return fmt.Sprintf("%s (%s)", currentM[1], strings.Join(labels, ", "))
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
