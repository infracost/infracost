package output

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"
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
	Projects             Projects         `json:"projects"`
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

// ToSchemaProject generates a schema.Project from a Project. The created schema.Project is not suitable to be
// used outside simple schema.Project to schema.Project comparisons. It contains missing information
// that cannot be inferred from a Project.
func (p Project) ToSchemaProject() *schema.Project {
	var pastResources []*schema.Resource
	if p.PastBreakdown != nil {
		pastResources = convertOutputResources(p.PastBreakdown.Resources)
	}

	var resources []*schema.Resource
	if p.Breakdown != nil {
		resources = convertOutputResources(p.Breakdown.Resources)
	}

	return &schema.Project{
		Name:          p.Name,
		Metadata:      p.Metadata,
		PastResources: pastResources,
		Resources:     resources,
	}
}

func convertOutputResources(outResources []Resource) []*schema.Resource {
	resources := make([]*schema.Resource, len(outResources))

	for i, resource := range outResources {
		resources[i] = &schema.Resource{
			Name:           resource.Name,
			CostComponents: convertCostComponents(resource.CostComponents),
			SubResources:   convertOutputResources(resource.SubResources),
			HourlyCost:     resource.HourlyCost,
			MonthlyCost:    resource.MonthlyCost,
			ResourceType:   resource.ResourceType(),
		}
	}

	return resources
}

func convertCostComponents(outComponents []CostComponent) []*schema.CostComponent {
	components := make([]*schema.CostComponent, len(outComponents))

	for i, c := range outComponents {
		sc := &schema.CostComponent{
			Name:            c.Name,
			Unit:            c.Unit,
			UnitMultiplier:  decimal.NewFromInt(1),
			HourlyCost:      c.HourlyCost,
			MonthlyCost:     c.MonthlyCost,
			HourlyQuantity:  c.HourlyQuantity,
			MonthlyQuantity: c.MonthlyQuantity,
		}
		sc.SetPrice(c.Price)

		components[i] = sc
	}

	return components
}

type Projects []Project

var exampleProjectsRegex = regexp.MustCompile(`^infracost\/(infracost\/examples|example-terraform)\/`)

func (r *Root) ExampleProjectName() string {
	if len(r.Projects) == 0 {
		return ""
	}

	for _, p := range r.Projects {
		if !exampleProjectsRegex.MatchString(p.Name) {
			return ""
		}
	}

	return r.Projects[0].Name
}

// Label returns the display name of the project
func (p *Project) Label() string {
	return p.Name
}

// LabelWithMetadata returns the display name of the project appended with any distinguishing
// metadata (Module path or Workspace)
func (p *Project) LabelWithMetadata() string {
	metadataInfo := []string{}
	if p.Metadata.TerraformModulePath != "" {
		metadataInfo = append(metadataInfo, "Module path: "+p.Metadata.TerraformModulePath)
	}
	if p.Metadata.WorkspaceLabel() != "" {
		metadataInfo = append(metadataInfo, "Workspace: "+p.Metadata.WorkspaceLabel())
	}

	if len(metadataInfo) == 0 {
		return p.Name
	}

	return fmt.Sprintf("%s (%s)", p.Name, strings.Join(metadataInfo, ", "))
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
	Name           string                 `json:"name"`
	Tags           map[string]string      `json:"tags,omitempty"`
	Metadata       map[string]interface{} `json:"metadata"`
	HourlyCost     *decimal.Decimal       `json:"hourlyCost"`
	MonthlyCost    *decimal.Decimal       `json:"monthlyCost"`
	CostComponents []CostComponent        `json:"costComponents,omitempty"`
	SubResources   []Resource             `json:"subresources,omitempty"`
}

func (r Resource) ResourceType() string {
	pieces := strings.Split(r.Name, ".")

	if len(pieces) >= 2 {
		return pieces[len(pieces)-2]
	}

	return r.Name
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
	DashboardEndpoint string
	NoColor           bool
	ShowSkipped       bool
	Fields            []string
	IncludeHTML       bool
	PolicyChecks      PolicyCheck
}

// PolicyCheck holds information if a given run has any policy checks enabled.
// This struct is used in templates to create useful cost policy outputs.
type PolicyCheck struct {
	Enabled  bool
	Failures PolicyCheckFailures
	Passed   []string
}

// HasFailed returns if the PolicyCheck has any cost policy failures
func (p PolicyCheck) HasFailed() bool {
	return len(p.Failures) > 0
}

// PolicyCheckFailures defines a list of policy check failures that can be collected from a policy evaluation.
type PolicyCheckFailures []string

// Error implements the Error interface returning the failures as a single message that can be used in stderr.
func (p PolicyCheckFailures) Error() string {
	if len(p) == 0 {
		return ""
	}

	out := bytes.NewBuffer([]byte("Policy check failed:\n\n"))

	for _, e := range p {
		out.WriteString(e + "\n")
	}

	return out.String()
}

type MarkdownOptions struct {
	WillUpdate          bool
	WillReplace         bool
	IncludeFeedbackLink bool
	OmitDetails         bool
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

	metadata := make(map[string]interface{})
	if r.Metadata != nil {
		for k, v := range r.Metadata {
			metadata[k] = v.Value()
		}
	}

	return Resource{
		Name:           r.Name,
		Metadata:       metadata,
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
		TimeGenerated:        time.Now().UTC(),
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

	seeDetailsMessage := ", rerun with --show-skipped to see details"

	if *r.Summary.TotalDetectedResources == 0 {
		msg += "No cloud resources were detected"
		return msg
	} else if *r.Summary.TotalDetectedResources == 1 {
		msg += "1 cloud resource was detected"
	} else {
		msg += fmt.Sprintf("%d cloud resources were detected", *r.Summary.TotalDetectedResources)
	}

	msg += ":"

	// Always show the total estimated, even if it's zero
	if r.Summary.TotalSupportedResources != nil {
		if *r.Summary.TotalSupportedResources == 1 {
			msg += "\n∙ 1 was estimated"
		} else {
			msg += fmt.Sprintf("\n∙ %d were estimated", *r.Summary.TotalSupportedResources)
		}

		allUsageBased := *r.Summary.TotalUsageBasedResources == *r.Summary.TotalSupportedResources

		if r.Summary.TotalUsageBasedResources != nil && *r.Summary.TotalUsageBasedResources > 0 {
			usageBasedCount := "1 of which"
			if allUsageBased {
				usageBasedCount = "it includes"
			}

			if *r.Summary.TotalUsageBasedResources > 1 {
				usageBasedCount = fmt.Sprintf("%d of which include", *r.Summary.TotalUsageBasedResources)
				if allUsageBased {
					usageBasedCount = "all of which include"
				}
			}

			msg += fmt.Sprintf(", %s usage-based costs, see %s", usageBasedCount, ui.SecondaryLinkString("https://infracost.io/usage-file"))
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
		} else {
			msg += seeDetailsMessage
		}
	}

	if r.Summary.TotalUnsupportedResources != nil && *r.Summary.TotalUnsupportedResources > 0 {
		count := "1 is"
		if *r.Summary.TotalUnsupportedResources > 1 {
			count = fmt.Sprintf("%d are", *r.Summary.TotalUnsupportedResources)
		}
		msg += fmt.Sprintf("\n∙ %s not supported yet", count)

		if showSkipped {
			msg += fmt.Sprintf(", see %s:", ui.SecondaryLinkString("https://infracost.io/requested-resources"))
			msg += formatCounts(r.Summary.UnsupportedResourceCounts)
		} else {
			msg += seeDetailsMessage
		}
	}

	if r.ShareURL != "" {
		msg += fmt.Sprintf("\n\nShare this cost estimate: %s", ui.LinkString(r.ShareURL))
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
		return resources[i].Metadata[groupKey].(string) < resources[j].Metadata[groupKey].(string)
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
