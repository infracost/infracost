package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/ui"
	"github.com/infracost/infracost/internal/usage"
)

var outputVersion = "0.2"

type Root struct {
	Version                   string           `json:"version"`
	Metadata                  Metadata         `json:"metadata"`
	RunID                     string           `json:"runId,omitempty"`
	ShareURL                  string           `json:"shareUrl,omitempty"`
	CloudURL                  string           `json:"cloudUrl,omitempty"`
	Currency                  string           `json:"currency"`
	Projects                  Projects         `json:"projects"`
	TotalHourlyCost           *decimal.Decimal `json:"totalHourlyCost"`
	TotalMonthlyCost          *decimal.Decimal `json:"totalMonthlyCost"`
	TotalMonthlyUsageCost     *decimal.Decimal `json:"totalMonthlyUsageCost,omitempty"`
	PastTotalHourlyCost       *decimal.Decimal `json:"pastTotalHourlyCost"`
	PastTotalMonthlyCost      *decimal.Decimal `json:"pastTotalMonthlyCost"`
	PastTotalMonthlyUsageCost *decimal.Decimal `json:"pastTotalMonthlyUsageCost,omitempty"`
	DiffTotalHourlyCost       *decimal.Decimal `json:"diffTotalHourlyCost"`
	DiffTotalMonthlyCost      *decimal.Decimal `json:"diffTotalMonthlyCost"`
	DiffTotalMonthlyUsageCost *decimal.Decimal `json:"diffTotalMonthlyUsageCost,omitempty"`
	TimeGenerated             time.Time        `json:"timeGenerated"`
	Summary                   *Summary         `json:"summary"`
	FullSummary               *Summary         `json:"-"`
	IsCIRun                   bool             `json:"-"`
	MissingPricesCount        int              `json:"missingPricesCount,omitempty"`
	MissingPricesComponents   []string         `json:"missingPricesComponents,omitempty"`
}

// HasUnsupportedResources returns if the summary has any unsupported resources.
// This is used to determine if the summary should be shown in different output
// formats.
func (r *Root) HasUnsupportedResources() bool {
	if r.Summary == nil {
		return false
	}

	if r.Summary.TotalUnsupportedResources == nil {
		return false
	}

	return *r.Summary.TotalUnsupportedResources > 0
}

type Project struct {
	Name          string                  `json:"name"`
	DisplayName   string                  `json:"displayName"`
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
func (p *Project) ToSchemaProject() *schema.Project {
	var pastResources []*schema.Resource
	if p.PastBreakdown != nil {
		pastResources = append(convertOutputResources(p.PastBreakdown.Resources, false), convertOutputResources(p.PastBreakdown.FreeResources, true)...)
	}

	var resources []*schema.Resource
	if p.Breakdown != nil {
		resources = append(convertOutputResources(p.Breakdown.Resources, false), convertOutputResources(p.Breakdown.FreeResources, true)...)
	}

	// clone the metadata to avoid unexpected effects from a shared pointer since
	// output.CompareTo may modify Metadata.PolicySha/PastPolicySha
	var clonedMetadata *schema.ProjectMetadata
	if p.Metadata != nil {
		m := *p.Metadata
		clonedMetadata = &m
	}

	return &schema.Project{
		Name:          p.Name,
		DisplayName:   p.DisplayName,
		Metadata:      clonedMetadata,
		PastResources: pastResources,
		Resources:     resources,
	}
}

func convertOutputResources(outResources []Resource, skip bool) []*schema.Resource {
	resources := make([]*schema.Resource, len(outResources))

	for i, resource := range outResources {

		var tagProp *schema.TagPropagation
		if resource.TagPropagation != nil {
			tagProp = &schema.TagPropagation{
				To:                    resource.TagPropagation.To,
				From:                  resource.TagPropagation.From,
				Tags:                  resource.TagPropagation.Tags,
				Attribute:             resource.TagPropagation.Attribute,
				HasRequiredAttributes: resource.TagPropagation.HasRequiredAttributes,
			}
		}

		resources[i] = &schema.Resource{
			Name:                                    resource.Name,
			IsSkipped:                               skip,
			Metadata:                                convertMetadata(resource.Metadata),
			CostComponents:                          convertCostComponents(resource.CostComponents),
			ActualCosts:                             convertActualCosts(resource.ActualCosts),
			SubResources:                            convertOutputResources(resource.SubResources, skip),
			Tags:                                    resource.Tags,
			DefaultTags:                             resource.DefaultTags,
			TagPropagation:                          tagProp,
			ProviderSupportsDefaultTags:             resource.ProviderSupportsDefaultTags,
			ProviderLink:                            resource.ProviderLink,
			HourlyCost:                              resource.HourlyCost,
			MonthlyCost:                             resource.MonthlyCost,
			MonthlyUsageCost:                        resource.MonthlyUsageCost,
			ResourceType:                            resource.ResourceType,
			MissingVarsCausingUnknownTagKeys:        resource.MissingVarsCausingUnknownTagKeys,
			MissingVarsCausingUnknownDefaultTagKeys: resource.MissingVarsCausingUnknownDefaultTagKeys,
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
			UsageBased:      c.UsageBased,
			PriceNotFound:   c.PriceNotFound,
		}
		sc.SetPrice(c.Price)

		components[i] = sc
	}

	return components
}

func convertActualCosts(outActualCosts []ActualCosts) []*schema.ActualCosts {
	actualCosts := make([]*schema.ActualCosts, len(outActualCosts))

	for i, ac := range outActualCosts {
		sac := &schema.ActualCosts{
			ResourceID:     ac.ResourceID,
			StartTimestamp: ac.StartTimestamp,
			EndTimestamp:   ac.EndTimestamp,
			CostComponents: convertCostComponents(ac.CostComponents),
		}

		actualCosts[i] = sac
	}

	return actualCosts
}

func convertMetadata(metadata map[string]interface{}) map[string]gjson.Result {
	result := make(map[string]gjson.Result)
	for k, v := range metadata {
		jsonBytes, err := json.Marshal(v)
		if err == nil {
			result[k] = gjson.ParseBytes(jsonBytes)
		}
	}

	return result
}

type Projects []Project

// IsRunQuotaExceeded checks if any of the project metadata have errored with a
// "run quota exceeded" error. If found it returns the associated message with
// this diag. This should be used when in any output that notifies the user.
func (projects Projects) IsRunQuotaExceeded() (string, bool) {
	for _, p := range projects {
		if msg, ok := p.Metadata.IsRunQuotaExceeded(); ok {
			return msg, true
		}
	}

	return "", false
}

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

// HasDiff returns true if any project has a difference in monthly cost or resources
func (r *Root) HasDiff() bool {
	for _, p := range r.Projects {
		if p.Diff != nil && (!p.Diff.TotalMonthlyCost.IsZero() || len(p.Diff.Resources) != 0) {
			return true
		}
	}

	return false
}

// Label returns the display name of the project
func (p *Project) Label() string {
	if p.DisplayName != "" {
		return p.DisplayName
	}

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
	Resources             []Resource       `json:"resources"`
	FreeResources         []Resource       `json:"freeResources,omitempty"`
	TotalHourlyCost       *decimal.Decimal `json:"totalHourlyCost"`
	TotalMonthlyCost      *decimal.Decimal `json:"totalMonthlyCost"`
	TotalMonthlyUsageCost *decimal.Decimal `json:"totalMonthlyUsageCost"`
}

// HasResources returns true if the breakdown has any resources or free resources.
// This is used to determine if the breakdown should be shown in the output.
func (b *Breakdown) HasResources() bool {
	return len(b.Resources) > 0 || len(b.FreeResources) > 0
}

func (b *Breakdown) TotalMonthlyBaselineCost() *decimal.Decimal {
	if b.TotalMonthlyCost == nil {
		return nil
	}

	if b.TotalMonthlyUsageCost == nil {
		return b.TotalMonthlyCost
	}

	return decimalPtr(b.TotalMonthlyCost.Sub(*b.TotalMonthlyUsageCost))
}

type CostComponent struct {
	Name            string           `json:"name"`
	Unit            string           `json:"unit"`
	HourlyQuantity  *decimal.Decimal `json:"hourlyQuantity"`
	MonthlyQuantity *decimal.Decimal `json:"monthlyQuantity"`
	Price           decimal.Decimal  `json:"price"`
	HourlyCost      *decimal.Decimal `json:"hourlyCost"`
	MonthlyCost     *decimal.Decimal `json:"monthlyCost"`
	UsageBased      bool             `json:"usageBased,omitempty"`
	PriceNotFound   bool             `json:"priceNotFound"`
}

type ActualCosts struct {
	ResourceID     string          `json:"resourceId"`
	StartTimestamp time.Time       `json:"startTimestamp"`
	EndTimestamp   time.Time       `json:"endTimestamp"`
	CostComponents []CostComponent `json:"costComponents,omitempty"`
}

type Resource struct {
	Name                                    string                 `json:"name"`
	ResourceType                            string                 `json:"resourceType,omitempty"`
	Tags                                    *map[string]string     `json:"tags,omitempty"`
	DefaultTags                             *map[string]string     `json:"defaultTags,omitempty"`
	TagPropagation                          *TagPropagation        `json:"tagPropagation,omitempty"`
	ProviderSupportsDefaultTags             bool                   `json:"providerSupportsDefaultTags,omitempty"`
	ProviderLink                            string                 `json:"providerLink,omitempty"`
	Metadata                                map[string]interface{} `json:"metadata"`
	HourlyCost                              *decimal.Decimal       `json:"hourlyCost,omitempty"`
	MonthlyCost                             *decimal.Decimal       `json:"monthlyCost,omitempty"`
	MonthlyUsageCost                        *decimal.Decimal       `json:"monthlyUsageCost,omitempty"`
	CostComponents                          []CostComponent        `json:"costComponents,omitempty"`
	ActualCosts                             []ActualCosts          `json:"actualCosts,omitempty"`
	SubResources                            []Resource             `json:"subresources,omitempty"`
	MissingVarsCausingUnknownTagKeys        []string               `json:"missingVarsCausingUnknownTagKeys,omitempty"`
	MissingVarsCausingUnknownDefaultTagKeys []string               `json:"missingVarsCausingUnknownDefaultTagKeys,omitempty"`
}

type TagPropagation struct {
	To                    string             `json:"to"`
	From                  *string            `json:"from,omitempty"`
	Tags                  *map[string]string `json:"tags,omitempty"`
	Attribute             string             `json:"attribute"`
	HasRequiredAttributes bool               `json:"hasRequiredAttributes,omitempty"`
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
	ShowAllProjects   bool
	ShowOnlyChanges   bool
	Fields            []string
	IncludeHTML       bool
	PolicyOutput      PolicyOutput
	diffMsg           string
	originalSize      int
	CurrencyFormat    string
}

// PolicyOutput holds normalized PolicyCheck and TagPolicyCheck data so it can be output in
// a uniform "Policies" section of the infracost comment.
type PolicyOutput struct {
	HasFailures bool
	HasWarnings bool
	Checks      []PolicyCheckOutput
}

type PolicyCheckOutput struct {
	Name            string
	Failure         bool
	Warning         bool
	Message         string
	Details         []string
	ResourceDetails []PolicyCheckResourceDetails
	TruncatedCount  int
}

type PolicyCheckResourceDetails struct {
	Address      string
	ResourceType string
	Path         string
	Line         int
	Violations   []PolicyCheckViolations
}

type PolicyCheckViolations struct {
	Details      []string
	ProjectNames []string
}

// NewPolicyOutput normalizes a PolicyCheck suitable
// for use in the output markdown template.
func NewPolicyOutput(pc PolicyCheck) PolicyOutput {
	po := PolicyOutput{}

	if pc.Enabled && len(pc.Failures) > 0 {
		po.HasFailures = true
		po.Checks = append(po.Checks, PolicyCheckOutput{
			Name:    "Cost policy failed",
			Failure: true,
			Details: pc.Failures,
		})
	}

	if pc.Enabled && len(pc.Passed) > 0 {
		po.Checks = append(po.Checks, PolicyCheckOutput{
			Name:    "Cost policy passed",
			Details: pc.Passed,
		})
	}

	return po
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

	out := &strings.Builder{}
	out.WriteString("Policy check failed:\n\n")

	for _, e := range p {
		out.WriteString(" - " + e + "\n")
	}

	return out.String()
}

// LoadCommentData reads the file at the path into a string.
func LoadCommentData(path string) (string, error) {
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", errors.New("comment data path does not exist ")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("error reading comment data file %w", err)
	}

	return string(data), nil
}

// GovernanceFailures defines a list of governance failures that were returned from infracost cloud.
type GovernanceFailures []string

// Error implements the Error interface returning the failures as a single message that can be used in stderr.
func (g GovernanceFailures) Error() string {
	if len(g) == 0 {
		return ""
	}

	out := &strings.Builder{}
	out.WriteString("Governance check failed:\n\n")

	for _, f := range g {
		out.WriteString(fmt.Sprintf(" - %s\n", f))
	}

	return out.String()
}

type MarkdownOptions struct {
	WillUpdate          bool
	WillReplace         bool
	IncludeFeedbackLink bool
	OmitDetails         bool
	BasicSyntax         bool
	MaxMessageSize      int
	Additional          string
}

func outputBreakdown(c *config.Config, resources []*schema.Resource) *Breakdown {
	supportedResources := make([]Resource, 0, len(resources))
	freeResources := make([]Resource, 0, len(resources))

	for _, r := range resources {
		if r.IsSkipped {
			if c.TagPoliciesEnabled && r.Tags != nil {
				freeResources = append(freeResources, newResource(r, nil, nil, nil))
			}

			continue
		}
		supportedResources = append(supportedResources, outputResource(r))
	}

	sortResources(supportedResources, "")
	sortResources(freeResources, "")

	totalHourlyCost, totalMonthlyCost, totalMonthlyUsageCost := calculateTotalCosts(supportedResources)

	return &Breakdown{
		Resources:             supportedResources,
		FreeResources:         freeResources,
		TotalHourlyCost:       totalHourlyCost,
		TotalMonthlyCost:      totalMonthlyCost,
		TotalMonthlyUsageCost: totalMonthlyUsageCost,
	}
}
func outputResource(r *schema.Resource) Resource {
	comps := outputCostComponents(r.CostComponents)

	actualCosts := outputActualCosts(r.ActualCosts)

	subresources := make([]Resource, 0, len(r.SubResources))
	for _, s := range r.SubResources {
		subresources = append(subresources, outputResource(s))
	}

	return newResource(r, comps, actualCosts, subresources)
}

func newResource(r *schema.Resource, comps []CostComponent, actualCosts []ActualCosts, subresources []Resource) Resource {
	metadata := make(map[string]interface{})
	for k, v := range r.Metadata {
		metadata[k] = v.Value()
	}

	var tagProp *TagPropagation
	if r.TagPropagation != nil {
		tagProp = &TagPropagation{
			To:                    r.TagPropagation.To,
			From:                  r.TagPropagation.From,
			Tags:                  r.TagPropagation.Tags,
			Attribute:             r.TagPropagation.Attribute,
			HasRequiredAttributes: r.TagPropagation.HasRequiredAttributes,
		}
	}

	return Resource{
		Name:                                    r.Name,
		ResourceType:                            r.ResourceType,
		Metadata:                                metadata,
		Tags:                                    r.Tags,
		DefaultTags:                             r.DefaultTags,
		TagPropagation:                          tagProp,
		ProviderSupportsDefaultTags:             r.ProviderSupportsDefaultTags,
		ProviderLink:                            r.ProviderLink,
		HourlyCost:                              r.HourlyCost,
		MonthlyCost:                             r.MonthlyCost,
		MonthlyUsageCost:                        r.MonthlyUsageCost,
		CostComponents:                          comps,
		ActualCosts:                             actualCosts,
		SubResources:                            subresources,
		MissingVarsCausingUnknownTagKeys:        r.MissingVarsCausingUnknownTagKeys,
		MissingVarsCausingUnknownDefaultTagKeys: r.MissingVarsCausingUnknownDefaultTagKeys,
	}
}

func outputCostComponents(costComponents []*schema.CostComponent) []CostComponent {
	comps := make([]CostComponent, 0, len(costComponents))
	for _, c := range costComponents {
		comps = append(comps, CostComponent{
			Name:            c.Name,
			Unit:            c.Unit,
			HourlyQuantity:  c.UnitMultiplierHourlyQuantity(),
			MonthlyQuantity: c.UnitMultiplierMonthlyQuantity(),
			Price:           c.UnitMultiplierPrice(),
			HourlyCost:      c.HourlyCost,
			MonthlyCost:     c.MonthlyCost,
			UsageBased:      c.UsageBased,
			PriceNotFound:   c.PriceNotFound,
		})
	}
	return comps
}

func outputActualCosts(actualCosts []*schema.ActualCosts) []ActualCosts {
	acs := make([]ActualCosts, 0, len(actualCosts))
	for _, ac := range actualCosts {
		acs = append(acs, ActualCosts{
			ResourceID:     ac.ResourceID,
			StartTimestamp: ac.StartTimestamp,
			EndTimestamp:   ac.EndTimestamp,
			CostComponents: outputCostComponents(ac.CostComponents),
		})
	}
	return acs
}

func ToOutputFormat(c *config.Config, projects []*schema.Project) (Root, error) {
	var totalMonthlyCost, totalHourlyCost, totalMonthlyUsageCost,
		pastTotalMonthlyCost, pastTotalHourlyCost, pastTotalMonthlyUsageCost,
		diffTotalMonthlyCost, diffTotalHourlyCost, diffTotalMonthlyUsageCost *decimal.Decimal

	outProjects := make([]Project, 0, len(projects))
	summaries := make([]*Summary, 0, len(projects))
	fullSummaries := make([]*Summary, 0, len(projects))

	for _, project := range projects {
		var pastBreakdown, breakdown, diff *Breakdown

		breakdown = outputBreakdown(c, project.Resources)

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

			if breakdown.TotalMonthlyUsageCost != nil {
				if totalMonthlyUsageCost == nil {
					totalMonthlyUsageCost = decimalPtr(decimal.Zero)
				}
				totalMonthlyUsageCost = decimalPtr(totalMonthlyUsageCost.Add(*breakdown.TotalMonthlyUsageCost))
			}
		}

		if project.HasDiff {
			pastBreakdown = outputBreakdown(c, project.PastResources)
			diff = outputBreakdown(c, project.Diff)

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

				if pastBreakdown.TotalMonthlyUsageCost != nil {
					if pastTotalMonthlyUsageCost == nil {
						pastTotalMonthlyUsageCost = decimalPtr(decimal.Zero)
					}
					pastTotalMonthlyUsageCost = decimalPtr(pastTotalMonthlyUsageCost.Add(*pastBreakdown.TotalMonthlyUsageCost))
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

				if diff.TotalMonthlyUsageCost != nil {
					if diffTotalMonthlyUsageCost == nil {
						diffTotalMonthlyUsageCost = decimalPtr(decimal.Zero)
					}
					diffTotalMonthlyUsageCost = decimalPtr(diffTotalMonthlyUsageCost.Add(*diff.TotalMonthlyUsageCost))
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
			DisplayName:   project.DisplayName,
			Metadata:      project.Metadata,
			PastBreakdown: pastBreakdown,
			Breakdown:     breakdown,
			Diff:          diff,
			Summary:       summary,
			fullSummary:   fullSummary,
		})
	}

	out := Root{
		Version:                   outputVersion,
		Projects:                  outProjects,
		TotalHourlyCost:           totalHourlyCost,
		TotalMonthlyCost:          totalMonthlyCost,
		TotalMonthlyUsageCost:     totalMonthlyUsageCost,
		PastTotalHourlyCost:       pastTotalHourlyCost,
		PastTotalMonthlyCost:      pastTotalMonthlyCost,
		PastTotalMonthlyUsageCost: pastTotalMonthlyUsageCost,
		DiffTotalHourlyCost:       diffTotalHourlyCost,
		DiffTotalMonthlyCost:      diffTotalMonthlyCost,
		DiffTotalMonthlyUsageCost: diffTotalMonthlyUsageCost,
		TimeGenerated:             time.Now().UTC(),
		Summary:                   MergeSummaries(summaries),
		FullSummary:               MergeSummaries(fullSummaries),
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
	}

	if r.Summary.TotalNoPriceResources != nil && *r.Summary.TotalNoPriceResources > 0 {
		if *r.Summary.TotalNoPriceResources == 1 {
			msg += "\n∙ 1 was free"
		} else {
			msg += fmt.Sprintf("\n∙ %d were free", *r.Summary.TotalNoPriceResources)
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

	// Add missing prices warning - this is the key fix for the bug
	if r.MissingPricesCount > 0 {
		warningMsg := ui.WarningString("WARNING:")
		if r.MissingPricesCount == 1 {
			msg += fmt.Sprintf("\n\n%s 1 price missing, costs may be incomplete", warningMsg)
		} else {
			msg += fmt.Sprintf("\n\n%s %d prices missing, costs may be incomplete", warningMsg, r.MissingPricesCount)
		}
		
		if showSkipped && len(r.MissingPricesComponents) > 0 {
			msg += "\nMissing prices for:"
			for _, component := range r.MissingPricesComponents {
				msg += fmt.Sprintf("\n  ∙ %s", component)
			}
		} else if len(r.MissingPricesComponents) > 0 {
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

func hasSupportedProvider(rType string) bool {
	return strings.HasPrefix(rType, "aws_") || // tf
		strings.HasPrefix(rType, "google_") || // tf
		strings.HasPrefix(rType, "azurerm_") || // tf
		strings.HasPrefix(rType, "AWS::") // cf
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
		if !opts.IncludeUnsupportedProviders && !hasSupportedProvider(r.ResourceType) {
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

func calculateTotalCosts(resources []Resource) (*decimal.Decimal, *decimal.Decimal, *decimal.Decimal) {
	totalHourlyCost := decimalPtr(decimal.Zero)
	totalMonthlyCost := decimalPtr(decimal.Zero)
	totalMonthlyUsageCost := decimalPtr(decimal.Zero)

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

		if r.MonthlyUsageCost != nil {
			if totalMonthlyUsageCost == nil {
				totalMonthlyUsageCost = decimalPtr(decimal.Zero)
			}

			totalMonthlyUsageCost = decimalPtr(totalMonthlyUsageCost.Add(*r.MonthlyUsageCost))

		}

	}

	return totalHourlyCost, totalMonthlyCost, totalMonthlyUsageCost
}

func sortResources(resources []Resource, groupKey string) {
	sort.Slice(resources, func(i, j int) bool {
		// if they are in different groups, sort by group name
		if groupKey != "" && resources[i].Metadata[groupKey] != resources[j].Metadata[groupKey] {
			return resources[i].Metadata[groupKey].(string) < resources[j].Metadata[groupKey].(string)
		}

		// if the costs are different, sort by cost descending
		if resources[i].MonthlyCost == nil {
			if resources[j].MonthlyCost != nil {
				return false
			}
		} else if resources[j].MonthlyCost == nil {
			return true
		} else if !resources[i].MonthlyCost.Equal(*resources[j].MonthlyCost) {
			return resources[i].MonthlyCost.GreaterThan(*resources[j].MonthlyCost)
		}

		// Sort by name
		return resources[i].Name < resources[j].Name
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

func usageCostsEnabled(out Root) bool {
	return out.Metadata.UsageApiEnabled || out.Metadata.UsageFilePath != "" || out.Metadata.ConfigFileHasUsageFile
}
