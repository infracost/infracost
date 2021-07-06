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

		summaries = append(summaries, input.Root.Summary)

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
	}

	combined.Version = outputVersion
	combined.Projects = projects
	combined.TotalHourlyCost = totalHourlyCost
	combined.TotalMonthlyCost = totalMonthlyCost
	combined.TimeGenerated = time.Now()
	combined.Summary = MergeSummaries(summaries)

	return combined
}
