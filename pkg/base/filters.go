package base

type Filter struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Operation string `json:"operation,omitempty"`
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
