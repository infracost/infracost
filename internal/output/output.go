package output

import (
	"fmt"
	"sort"
	"time"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/providers/terraform"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"
	"github.com/infracost/infracost/internal/usage"
)

var outputVersion = "0.2"

type Root struct {
	Version              string           `json:"version"`
	RunID                string           `json:"runId,omitempty"`
	ShareURL             string           `json:"shareUrl,omitempty"`
	Currency             string           `json:"currency"`
	Projects             []Project        `json:"projects"`
	TotalHourlyCost      *decimal.Decimal `json:"totalHourlyCost"`
	TotalMonthlyCost     *decimal.Decimal `json:"totalMonthlyCost"`
	PastTotalHourlyCost  *decimal.Decimal `json:"pastTotalHourlyCost"`
	PastTotalMonthlyCost *decimal.Decimal `json:"pastTotalMonthlyCost"`
	DiffTotalHourlyCost  *decimal.Decimal `json:"diffTotalHourlyCost"`
	DiffTotalMonthlyCost *decimal.Decimal `json:"diffTotalMonthlyCost"`
	TimeGenerated        time.Time        `json:"timeGenerated"`
	Summary              *Summary         `json:"summary"`
	FullSummary          *Summary         `json:"-"`
	IsCIRun              bool             `json:"-"`
}

type Project struct {
	Name          string                  `json:"name"`
	Metadata      *schema.ProjectMetadata `json:"metadata"`
	PastBreakdown *Breakdown              `json:"pastBreakdown"`
	Breakdown     *Breakdown              `json:"breakdown"`
	Diff          *Breakdown              `json:"diff"`
	Summary       *Summary                `json:"summary"`
	fullSummary   *Summary
}

func (p *Project) Label(dashboardEnabled bool) string {
	if !dashboardEnabled {
		return p.Name
	}
	return fmt.Sprintf("%s (%s)", p.Name, p.Metadata.Path)
}

type Breakdown struct {
	Resources        []Resource       `json:"resources"`
	TotalHourlyCost  *decimal.Decimal `json:"totalHourlyCost"`
	TotalMonthlyCost *decimal.Decimal `json:"totalMonthlyCost"`
}

type CostComponent struct {
	Name            string           `json:"name"`
	Unit            string           `json:"unit"`
	HourlyQuantity  *decimal.Decimal `json:"hourlyQuantity"`
	MonthlyQuantity *decimal.Decimal `json:"monthlyQuantity"`
	Price           decimal.Decimal  `json:"price"`
	HourlyCost      *decimal.Decimal `json:"hourlyCost"`
	MonthlyCost     *decimal.Decimal `json:"monthlyCost"`
}

type Resource struct {
	Name           string            `json:"name"`
	Tags           map[string]string `json:"tags,omitempty"`
	Metadata       map[string]string `json:"metadata"`
	HourlyCost     *decimal.Decimal  `json:"hourlyCost"`
	MonthlyCost    *decimal.Decimal  `json:"monthlyCost"`
	CostComponents []CostComponent   `json:"costComponents,omitempty"`
	SubResources   []Resource        `json:"subresources,omitempty"`
}

type Summary struct {
	TotalResources            *int `json:"totalResources,omitempty"`
	TotalDetectedResources    *int `json:"totalDetectedResources,omitempty"`
	TotalSupportedResources   *int `json:"totalSupportedResources,omitempty"`
	TotalUnsupportedResources *int `json:"totalUnsupportedResources,omitempty"`
	TotalUsageBasedResources  *int `json:"totalUsageBasedResources,omitempty"`
	TotalNoPriceResources     *int `json:"totalNoPriceResources,omitempty"`

	SupportedResourceCounts   *map[string]int `json:"supportedResourceCounts,omitempty"`
	UnsupportedResourceCounts *map[string]int `json:"unsupportedResourceCounts,omitempty"`
	NoPriceResourceCounts     *map[string]int `json:"noPriceResourceCounts,omitempty"`

	EstimatedUsageCounts   *map[string]int `json:"-"`
	UnestimatedUsageCounts *map[string]int `json:"-"`
	TotalEstimatedUsages   *int            `json:"-"`
	TotalUnestimatedUsages *int            `json:"-"`
}

type SummaryOptions struct {
	IncludeUnsupportedProviders bool
	OnlyFields                  []string
}

type Options struct {
	DashboardEnabled bool
	NoColor          bool
	ShowSkipped      bool
	Fields           []string
}

type MarkdownOptions struct {
	WillUpdate          bool
	WillReplace         bool
	IncludeFeedbackLink bool
	BasicSyntax         bool
}

func outputBreakdown(resources []*schema.Resource) *Breakdown {
	arr := make([]Resource, 0, len(resources))

	for _, r := range resources {
		if r.IsSkipped {
			continue
		}
		arr = append(arr, outputResource(r))
	}

	sortResources(arr, "")

	totalMonthlyCost, totalHourlyCost := calculateTotalCosts(arr)

	return &Breakdown{
		Resources:        arr,
		TotalHourlyCost:  totalMonthlyCost,
		TotalMonthlyCost: totalHourlyCost,
	}
}

func outputResource(r *schema.Resource) Resource {
	comps := make([]CostComponent, 0, len(r.CostComponents))
	for _, c := range r.CostComponents {

		comps = append(comps, CostComponent{
			Name:            c.Name,
			Unit:            c.Unit,
			HourlyQuantity:  c.UnitMultiplierHourlyQuantity(),
			MonthlyQuantity: c.UnitMultiplierMonthlyQuantity(),
			Price:           c.UnitMultiplierPrice(),
			HourlyCost:      c.HourlyCost,
			MonthlyCost:     c.MonthlyCost,
		})
	}

	subresources := make([]Resource, 0, len(r.SubResources))
	for _, s := range r.SubResources {
		subresources = append(subresources, outputResource(s))
	}

	return Resource{
		Name:           r.Name,
		Metadata:       map[string]string{},
		Tags:           r.Tags,
		HourlyCost:     r.HourlyCost,
		MonthlyCost:    r.MonthlyCost,
		CostComponents: comps,
		SubResources:   subresources,
	}
}

func ToOutputFormat(projects []*schema.Project) (Root, error) {
	var totalMonthlyCost, totalHourlyCost,
		pastTotalMonthlyCost, pastTotalHourlyCost,
		diffTotalMonthlyCost, diffTotalHourlyCost *decimal.Decimal

	outProjects := make([]Project, 0, len(projects))
	summaries := make([]*Summary, 0, len(projects))
	fullSummaries := make([]*Summary, 0, len(projects))

	for _, project := range projects {
		var pastBreakdown, breakdown, diff *Breakdown

		breakdown = outputBreakdown(project.Resources)

		if breakdown != nil {
			if breakdown.TotalHourlyCost != nil {
				if totalHourlyCost == nil {
					totalHourlyCost = decimalPtr(decimal.Zero)
				}
				totalHourlyCost = decimalPtr(totalHourlyCost.Add(*breakdown.TotalHourlyCost))
			}

			if breakdown.TotalMonthlyCost != nil {
				if totalMonthlyCost == nil {
					totalMonthlyCost = decimalPtr(decimal.Zero)
				}
				totalMonthlyCost = decimalPtr(totalMonthlyCost.Add(*breakdown.TotalMonthlyCost))
			}
		}

		if project.HasDiff {
			pastBreakdown = outputBreakdown(project.PastResources)
			diff = outputBreakdown(project.Diff)

			if pastBreakdown != nil {
				if pastBreakdown.TotalHourlyCost != nil {
					if pastTotalHourlyCost == nil {
						pastTotalHourlyCost = decimalPtr(decimal.Zero)
					}
					pastTotalHourlyCost = decimalPtr(pastTotalHourlyCost.Add(*pastBreakdown.TotalHourlyCost))
				}

				if pastBreakdown.TotalMonthlyCost != nil {
					if pastTotalMonthlyCost == nil {
						pastTotalMonthlyCost = decimalPtr(decimal.Zero)
					}
					pastTotalMonthlyCost = decimalPtr(pastTotalMonthlyCost.Add(*pastBreakdown.TotalMonthlyCost))
				}
			}

			if diff != nil {
				if diff.TotalHourlyCost != nil {
					if diffTotalHourlyCost == nil {
						diffTotalHourlyCost = decimalPtr(decimal.Zero)
					}
					diffTotalHourlyCost = decimalPtr(diffTotalHourlyCost.Add(*diff.TotalHourlyCost))
				}

				if diff.TotalMonthlyCost != nil {
					if diffTotalMonthlyCost == nil {
						diffTotalMonthlyCost = decimalPtr(decimal.Zero)
					}
					diffTotalMonthlyCost = decimalPtr(diffTotalMonthlyCost.Add(*diff.TotalMonthlyCost))
				}
			}
		}

		summary, err := BuildSummary(project.Resources, SummaryOptions{
			OnlyFields: []string{
				"TotalDetectedResources",
				"TotalSupportedResources",
				"TotalUnsupportedResources",
				"TotalUsageBasedResources",
				"TotalNoPriceResources",
				"UnsupportedResourceCounts",
				"NoPriceResourceCounts",
			},
		})
		if err != nil {
			return Root{}, err
		}
		summaries = append(summaries, summary)

		fullSummary, err := BuildSummary(project.Resources, SummaryOptions{IncludeUnsupportedProviders: true})
		if err != nil {
			return Root{}, err
		}
		fullSummaries = append(fullSummaries, fullSummary)

		outProjects = append(outProjects, Project{
			Name:          project.Name,
			Metadata:      project.Metadata,
			PastBreakdown: pastBreakdown,
			Breakdown:     breakdown,
			Diff:          diff,
			Summary:       summary,
			fullSummary:   fullSummary,
		})
	}

	out := Root{
		Version:              outputVersion,
		Projects:             outProjects,
		TotalHourlyCost:      totalHourlyCost,
		TotalMonthlyCost:     totalMonthlyCost,
		PastTotalHourlyCost:  pastTotalHourlyCost,
		PastTotalMonthlyCost: pastTotalMonthlyCost,
		DiffTotalHourlyCost:  diffTotalHourlyCost,
		DiffTotalMonthlyCost: diffTotalMonthlyCost,
		TimeGenerated:        time.Now(),
		Summary:              MergeSummaries(summaries),
		FullSummary:          MergeSummaries(fullSummaries),
	}

	return out, nil
}

func (r *Root) summaryMessage(showSkipped bool) string {
	msg := ""

	if r.Summary == nil || r.Summary.TotalDetectedResources == nil {
		return msg
	}

	if *r.Summary.TotalDetectedResources == 0 {
		msg += "No cloud resources were detected"
		return msg
	} else if *r.Summary.TotalDetectedResources == 1 {
		msg += "1 cloud resource was detected"
	} else {
		msg += fmt.Sprintf("%d cloud resources were detected", *r.Summary.TotalDetectedResources)
	}

	if !showSkipped {
		msg += ", rerun with --show-skipped to see details"
	}

	msg += ":"

	// Always show the total estimated, even if it's zero
	if r.Summary.TotalSupportedResources != nil {
		if *r.Summary.TotalSupportedResources == 1 {
			msg += "\n∙ 1 was estimated"
		} else {
			msg += fmt.Sprintf("\n∙ %d were estimated", *r.Summary.TotalSupportedResources)
		}

		if r.Summary.TotalUsageBasedResources != nil && *r.Summary.TotalUsageBasedResources > 0 {
			if *r.Summary.TotalUsageBasedResources == 1 {
				msg += fmt.Sprintf(", 1 includes usage-based costs, see %s", "https://infracost.io/usage-file")
			} else {
				msg += fmt.Sprintf(", %d include usage-based costs, see %s", *r.Summary.TotalUsageBasedResources, "https://infracost.io/usage-file")
			}
		}
	}

	if r.Summary.TotalUnsupportedResources != nil && *r.Summary.TotalUnsupportedResources > 0 {
		if *r.Summary.TotalUnsupportedResources == 1 {
			msg += fmt.Sprintf("\n∙ 1 wasn't estimated, report it in %s", "https://github.com/infracost/infracost")
		} else {
			msg += fmt.Sprintf("\n∙ %d weren't estimated, report them in %s", *r.Summary.TotalUnsupportedResources, "https://github.com/infracost/infracost")
		}

		if showSkipped {
			msg += ":"
			msg += formatCounts(r.Summary.UnsupportedResourceCounts)
		}
	}

	if r.Summary.TotalNoPriceResources != nil && *r.Summary.TotalNoPriceResources > 0 {
		if *r.Summary.TotalNoPriceResources == 1 {
			msg += "\n∙ 1 was free"
		} else {
			msg += fmt.Sprintf("\n∙ %d were free", *r.Summary.TotalNoPriceResources)
		}

		if showSkipped {
			msg += ":"
			msg += formatCounts(r.Summary.NoPriceResourceCounts)
		}
	}

	if r.ShareURL != "" {
		msg += fmt.Sprintf("\n\nShare this cost estimate: %s", ui.LinkString(r.ShareURL))
	}

	if !r.IsCIRun {
		msg += fmt.Sprintf("\n\nAdd cost estimates to your pull requests: %s", ui.LinkString("https://infracost.io/cicd"))
	}

	return msg
}

func formatCounts(countMap *map[string]int) string {
	msg := ""

	if countMap == nil {
		return msg
	}

	type structMap struct {
		key   string
		value int
	}
	m := []structMap{}

	for t, c := range *countMap {
		m = append(m, structMap{key: t, value: c})
	}

	sort.Slice(m, func(i, j int) bool {
		if m[i].value == m[j].value {
			return m[i].key < m[j].key
		}
		return m[i].value > m[j].value
	})

	for _, i := range m {
		msg += fmt.Sprintf("\n  ∙ %d x %s", i.value, i.key)
	}

	return msg
}

func BuildSummary(resources []*schema.Resource, opts SummaryOptions) (*Summary, error) {
	s := &Summary{}

	supportedResourceCounts := make(map[string]int)
	unsupportedResourceCounts := make(map[string]int)
	noPriceResourceCounts := make(map[string]int)
	totalDetectedResources := 0
	totalSupportedResources := 0
	totalUnsupportedResources := 0
	totalUsageBasedResources := 0
	totalNoPriceResources := 0

	estimatedUsageCounts := make(map[string]int)
	unestimatedUsageCounts := make(map[string]int)
	totalEstimatedUsages := 0
	totalUnestimatedUsages := 0

	refFile, err := usage.LoadReferenceFile()
	if err != nil {
		return s, err
	}

	for _, r := range resources {
		if !opts.IncludeUnsupportedProviders && !terraform.HasSupportedProvider(r.ResourceType) {
			continue
		}

		totalDetectedResources++

		if r.NoPrice {
			totalNoPriceResources++
			if _, ok := noPriceResourceCounts[r.ResourceType]; !ok {
				noPriceResourceCounts[r.ResourceType] = 0
			}
			noPriceResourceCounts[r.ResourceType]++
		} else if r.IsSkipped {
			totalUnsupportedResources++
			if _, ok := unsupportedResourceCounts[r.ResourceType]; !ok {
				unsupportedResourceCounts[r.ResourceType] = 0
			}
			unsupportedResourceCounts[r.ResourceType]++
		} else {
			totalSupportedResources++
			if _, ok := supportedResourceCounts[r.ResourceType]; !ok {
				supportedResourceCounts[r.ResourceType] = 0
			}
			supportedResourceCounts[r.ResourceType]++

			if refFile.FindMatchingResourceUsage(r.Name) != nil {
				totalUsageBasedResources++
			}
		}

		for usage, isEstimated := range r.EstimationSummary {
			k := r.ResourceType + "." + usage
			if isEstimated {
				totalEstimatedUsages++
				if _, ok := estimatedUsageCounts[k]; !ok {
					estimatedUsageCounts[k] = 0
				}
				estimatedUsageCounts[k]++
			} else {
				totalUnestimatedUsages++
				if _, ok := unestimatedUsageCounts[k]; !ok {
					unestimatedUsageCounts[k] = 0
				}
				unestimatedUsageCounts[k]++
			}
		}
	}

	totalResources := len(resources)

	if len(opts.OnlyFields) == 0 || contains(opts.OnlyFields, "TotalResources") {
		s.TotalResources = &totalResources
	}
	if len(opts.OnlyFields) == 0 || contains(opts.OnlyFields, "TotalDetectedResources") {
		s.TotalDetectedResources = &totalDetectedResources
	}
	if len(opts.OnlyFields) == 0 || contains(opts.OnlyFields, "TotalSupportedResources") {
		s.TotalSupportedResources = &totalSupportedResources
	}
	if len(opts.OnlyFields) == 0 || contains(opts.OnlyFields, "TotalUnsupportedResources") {
		s.TotalUnsupportedResources = &totalUnsupportedResources
	}
	if len(opts.OnlyFields) == 0 || contains(opts.OnlyFields, "TotalUsageBasedResources") {
		s.TotalUsageBasedResources = &totalUsageBasedResources
	}
	if len(opts.OnlyFields) == 0 || contains(opts.OnlyFields, "TotalNoPriceResources") {
		s.TotalNoPriceResources = &totalNoPriceResources
	}
	if len(opts.OnlyFields) == 0 || contains(opts.OnlyFields, "SupportedResourceCounts") {
		s.SupportedResourceCounts = &supportedResourceCounts
	}
	if len(opts.OnlyFields) == 0 || contains(opts.OnlyFields, "UnsupportedResourceCounts") {
		s.UnsupportedResourceCounts = &unsupportedResourceCounts
	}
	if len(opts.OnlyFields) == 0 || contains(opts.OnlyFields, "NoPriceResourceCounts") {
		s.NoPriceResourceCounts = &noPriceResourceCounts
	}

	if len(opts.OnlyFields) == 0 || contains(opts.OnlyFields, "EstimatedUsageCounts") {
		s.EstimatedUsageCounts = &estimatedUsageCounts
	}
	if len(opts.OnlyFields) == 0 || contains(opts.OnlyFields, "UnestimatedUsageCounts") {
		s.UnestimatedUsageCounts = &unestimatedUsageCounts
	}
	if len(opts.OnlyFields) == 0 || contains(opts.OnlyFields, "TotalEstimatedUsages") {
		s.TotalEstimatedUsages = &totalEstimatedUsages
	}
	if len(opts.OnlyFields) == 0 || contains(opts.OnlyFields, "TotalUnestimatedUsages") {
		s.TotalUnestimatedUsages = &totalUnestimatedUsages
	}

	return s, nil
}

func MergeSummaries(summaries []*Summary) *Summary {
	merged := &Summary{}

	for _, s := range summaries {
		if s == nil {
			continue
		}

		merged.TotalResources = addIntPtrs(merged.TotalResources, s.TotalResources)
		merged.TotalDetectedResources = addIntPtrs(merged.TotalDetectedResources, s.TotalDetectedResources)
		merged.TotalSupportedResources = addIntPtrs(merged.TotalSupportedResources, s.TotalSupportedResources)
		merged.TotalUnsupportedResources = addIntPtrs(merged.TotalUnsupportedResources, s.TotalUnsupportedResources)
		merged.TotalUsageBasedResources = addIntPtrs(merged.TotalUsageBasedResources, s.TotalUsageBasedResources)
		merged.TotalNoPriceResources = addIntPtrs(merged.TotalNoPriceResources, s.TotalNoPriceResources)
		merged.SupportedResourceCounts = mergeCounts(merged.SupportedResourceCounts, s.SupportedResourceCounts)
		merged.UnsupportedResourceCounts = mergeCounts(merged.UnsupportedResourceCounts, s.UnsupportedResourceCounts)
		merged.NoPriceResourceCounts = mergeCounts(merged.NoPriceResourceCounts, s.NoPriceResourceCounts)

		merged.EstimatedUsageCounts = mergeCounts(merged.EstimatedUsageCounts, s.EstimatedUsageCounts)
		merged.UnestimatedUsageCounts = mergeCounts(merged.UnestimatedUsageCounts, s.UnestimatedUsageCounts)
		merged.TotalEstimatedUsages = addIntPtrs(merged.TotalEstimatedUsages, s.TotalEstimatedUsages)
		merged.TotalUnestimatedUsages = addIntPtrs(merged.TotalUnestimatedUsages, s.TotalUnestimatedUsages)
	}

	return merged
}

func calculateTotalCosts(resources []Resource) (*decimal.Decimal, *decimal.Decimal) {
	totalHourlyCost := decimalPtr(decimal.Zero)
	totalMonthlyCost := decimalPtr(decimal.Zero)

	for _, r := range resources {
		if r.HourlyCost != nil {
			if totalHourlyCost == nil {
				totalHourlyCost = decimalPtr(decimal.Zero)
			}

			totalHourlyCost = decimalPtr(totalHourlyCost.Add(*r.HourlyCost))
		}
		if r.MonthlyCost != nil {
			if totalMonthlyCost == nil {
				totalMonthlyCost = decimalPtr(decimal.Zero)
			}

			totalMonthlyCost = decimalPtr(totalMonthlyCost.Add(*r.MonthlyCost))
		}

	}

	return totalHourlyCost, totalMonthlyCost
}

func sortResources(resources []Resource, groupKey string) {
	sort.Slice(resources, func(i, j int) bool {
		// If an empty group key is passed just sort by name
		if groupKey == "" {
			return resources[i].Name < resources[j].Name
		}

		// If the resources are in the same group then sort by name
		if resources[i].Metadata[groupKey] == resources[j].Metadata[groupKey] {
			return resources[i].Name < resources[j].Name
		}

		// Sort by the group key
		return resources[i].Metadata[groupKey] < resources[j].Metadata[groupKey]
	})
}

func contains(arr []string, e string) bool {
	for _, a := range arr {
		if a == e {
			return true
		}
	}
	return false
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

func mergeCounts(c1 *map[string]int, c2 *map[string]int) *map[string]int {
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
