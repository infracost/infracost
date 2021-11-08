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

func Combine(currency string, inputs []ReportInput, opts Options) Root {
	var combined Root

	var totalHourlyCost *decimal.Decimal
	var totalMonthlyCost *decimal.Decimal
	var pastTotalHourlyCost *decimal.Decimal
	var pastTotalMonthlyCost *decimal.Decimal
	var diffTotalHourlyCost *decimal.Decimal
	var diffTotalMonthlyCost *decimal.Decimal

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
		if input.Root.PastTotalMonthlyCost != nil {
			if pastTotalMonthlyCost == nil {
				pastTotalMonthlyCost = decimalPtr(decimal.Zero)
			}

			pastTotalMonthlyCost = decimalPtr(pastTotalMonthlyCost.Add(*input.Root.PastTotalMonthlyCost))
		}
		if input.Root.PastTotalMonthlyCost != nil {
			if pastTotalMonthlyCost == nil {
				pastTotalMonthlyCost = decimalPtr(decimal.Zero)
			}

			pastTotalMonthlyCost = decimalPtr(pastTotalMonthlyCost.Add(*input.Root.PastTotalMonthlyCost))
		}
		if input.Root.DiffTotalMonthlyCost != nil {
			if diffTotalMonthlyCost == nil {
				diffTotalMonthlyCost = decimalPtr(decimal.Zero)
			}

			diffTotalMonthlyCost = decimalPtr(diffTotalMonthlyCost.Add(*input.Root.DiffTotalMonthlyCost))
		}
		if input.Root.DiffTotalMonthlyCost != nil {
			if diffTotalMonthlyCost == nil {
				diffTotalMonthlyCost = decimalPtr(decimal.Zero)
			}

			diffTotalMonthlyCost = decimalPtr(diffTotalMonthlyCost.Add(*input.Root.DiffTotalMonthlyCost))
		}

	}

	combined.Version = outputVersion
	combined.Currency = currency
	combined.Projects = projects
	combined.TotalHourlyCost = totalHourlyCost
	combined.TotalMonthlyCost = totalMonthlyCost
	combined.PastTotalHourlyCost = pastTotalHourlyCost
	combined.PastTotalMonthlyCost = pastTotalMonthlyCost
	combined.DiffTotalHourlyCost = diffTotalHourlyCost
	combined.DiffTotalMonthlyCost = diffTotalMonthlyCost
	combined.TimeGenerated = time.Now()
	combined.Summary = MergeSummaries(summaries)

	return combined
}
