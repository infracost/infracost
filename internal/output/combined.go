package output

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

type ReportInput struct {
	Metadata map[string]string
	Root     Root
}

func Load(data []byte) (Root, error) {
	var out Root
	err := json.Unmarshal(data, &out)
	return out, err
}

func Combine(inputs []ReportInput, opts Options) Root {
	var combined Root

	var totalHourlyCost *decimal.Decimal
	var totalMonthlyCost *decimal.Decimal

	projects := make([]Project, 0)
	summaries := make([]*Summary, 0, len(inputs))

	for _, input := range inputs {

		projects = append(projects, input.Root.Projects...)

		for _, r := range input.Root.Resources {
			for k, v := range input.Metadata {
				if r.Metadata == nil {
					r.Metadata = make(map[string]string)
				}
				r.Metadata[k] = v
			}

			combined.Resources = append(combined.Resources, r)
		}

		if input.Root.TotalHourlyCost != nil {
			if totalHourlyCost == nil {
				totalHourlyCost = decimalPtr(decimal.Zero)
			}

			totalHourlyCost = decimalPtr(totalHourlyCost.Add(*input.Root.TotalHourlyCost))
		}
		if input.Root.TotalMonthlyCost != nil {
			if totalMonthlyCost == nil {
				totalMonthlyCost = decimalPtr(decimal.Zero)
			}

			totalMonthlyCost = decimalPtr(totalMonthlyCost.Add(*input.Root.TotalMonthlyCost))
		}

		summaries = append(summaries, input.Root.Summary)
	}

	sortResources(combined.Resources, opts.GroupKey)

	combined.Version = outputVersion
	combined.Projects = projects
	combined.TotalHourlyCost = totalHourlyCost
	combined.TotalMonthlyCost = totalMonthlyCost
	combined.TimeGenerated = time.Now()
	combined.Summary = combinedResourceSummaries(summaries)

	return combined
}

func combinedResourceSummaries(summaries []*Summary) *Summary {
	combined := &Summary{}

	for _, s := range summaries {
		if s == nil {
			continue
		}

		combined.SupportedResourceCounts = combineCounts(combined.SupportedResourceCounts, s.SupportedResourceCounts)
		combined.UnsupportedResourceCounts = combineCounts(combined.UnsupportedResourceCounts, s.UnsupportedResourceCounts)
		combined.TotalSupportedResources = addIntPtrs(combined.TotalSupportedResources, s.TotalSupportedResources)
		combined.TotalUnsupportedResources = addIntPtrs(combined.TotalUnsupportedResources, s.TotalUnsupportedResources)
		combined.TotalNoPriceResources = addIntPtrs(combined.TotalNoPriceResources, s.TotalNoPriceResources)
		combined.TotalResources = addIntPtrs(combined.TotalResources, s.TotalResources)
	}

	return combined
}

func combineCounts(c1 *map[string]int, c2 *map[string]int) *map[string]int {
	if c1 == nil && c2 == nil {
		return nil
	}

	res := make(map[string]int)

	if c1 != nil {
		for k, v := range *c1 {
			res[k] = v
		}
	}

	if c2 != nil {
		for k, v := range *c2 {
			res[k] += v
		}
	}

	return &res
}

func addIntPtrs(i1 *int, i2 *int) *int {
	if i1 == nil && i2 == nil {
		return nil
	}

	val1 := 0
	if i1 != nil {
		val1 = *i1
	}

	val2 := 0
	if i2 != nil {
		val2 = *i2
	}

	res := val1 + val2
	return &res
}
