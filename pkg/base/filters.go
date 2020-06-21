package base

import (
	"fmt"
	"sort"
)

type Filter struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Operation string `json:"operation,omitempty"`
}

type ValueMapping struct {
	FromKey   string
	ToKey     string
	ToValueFn func(fromVal interface{}) string
}

func (v *ValueMapping) MappedValue(fromVal interface{}) string {
	if v.ToValueFn != nil {
		return v.ToValueFn(fromVal)
	}
	if fromVal == nil {
		return ""
	}
	return fmt.Sprintf("%v", fromVal)
}

func MergeFilters(filtersets ...[]Filter) []Filter {
	m := map[string]Filter{}
	for _, filterset := range filtersets {
		for _, filter := range filterset {
			m[filter.Key] = filter
		}
	}

	mergedFilters := make([]Filter, 0, len(m))
	for _, filter := range m {
		mergedFilters = append(mergedFilters, filter)
	}
	return mergedFilters
}

func MapFilters(valueMappings []ValueMapping, values map[string]interface{}) []Filter {
	mappedFilters := []Filter{}
	for fromKey, fromVal := range values {
		var valueMapping *ValueMapping
		for _, v := range valueMappings {
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

func SortFiltersByKey(filters []Filter) []Filter {
	copied := append([]Filter{}, filters...)
	sort.Slice(copied, func(i, j int) bool {
		return copied[i].Key < copied[j].Key
	})
	return copied
}
