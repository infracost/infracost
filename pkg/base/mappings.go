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
	PriceMappings       map[string]*PriceMapping
	SubResourceMappings map[string]*ResourceMapping
}

type PriceMapping struct {
	TimeUnit       string
	DefaultFilters []Filter
	ValueMappings  []ValueMapping
	ShouldSkip     func(values map[string]interface{}) bool
	CalculateCost  func(price decimal.Decimal) decimal.Decimal
}

func (p *PriceMapping) GetFilters(values map[string]interface{}) []Filter {
	return MergeFilters(p.DefaultFilters, p.valueFilters(values))
}

func (p *PriceMapping) valueFilters(values map[string]interface{}) []Filter {
	mappedFilters := []Filter{}
	for fromKey, fromVal := range values {
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
