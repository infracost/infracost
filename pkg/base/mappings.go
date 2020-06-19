package base

import (
	"fmt"

	"github.com/shopspring/decimal"
)

type ValueMapping struct {
	FromKey   string
	ToKey     string
	ToValueFn func(fromVal interface{}) string
}

func (v *ValueMapping) MappedValue(fromVal interface{}) string {
	if v.ToValueFn != nil {
		return v.ToValueFn(fromVal)
	}
	return fmt.Sprintf("%v", fromVal)
}

type ResourceMapping struct {
	PriceMappings                map[string]*PriceMapping
	SubResourceMappings          map[string]*ResourceMapping
	OverrideSubResourceRawValues func(resource Resource) map[string][]interface{}
	AdjustCost                   func(resource Resource, cost decimal.Decimal) decimal.Decimal
	NonCostable                  bool
}

type PriceMapping struct {
	TimeUnit        string
	DefaultFilters  []Filter
	ValueMappings   []ValueMapping
	OverrideFilters func(resource Resource) []Filter
	ShouldSkip      func(values map[string]interface{}) bool
	CalculateCost   func(price decimal.Decimal, resource Resource) decimal.Decimal
}

func (p *PriceMapping) GetFilters(resource Resource) []Filter {
	overridenFilters := make([]Filter, 0)
	if p.OverrideFilters != nil {
		overridenFilters = p.OverrideFilters(resource)
	}
	return MergeFilters(p.DefaultFilters, p.valueFilters(resource), overridenFilters)
}

func (p *PriceMapping) valueFilters(resource Resource) []Filter {

	mappedFilters := []Filter{}
	for fromKey, fromVal := range resource.RawValues() {
		var valueMapping *ValueMapping
		for _, v := range p.ValueMappings {
			if v.FromKey == fromKey {
				valueMapping = &v
				break
			}
		}

		if valueMapping == nil {
			continue
		}

		toVal := valueMapping.MappedValue(fromVal)
		if toVal != "" {
			mappedFilters = append(mappedFilters, Filter{
				Key:   valueMapping.ToKey,
				Value: toVal,
			})
		}
	}
	return mappedFilters
}
