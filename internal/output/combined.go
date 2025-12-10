package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"golang.org/x/mod/semver"

	"github.com/infracost/infracost/internal/metrics"

	"github.com/infracost/infracost/internal/clierror"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"
)

var (
	minOutputVersion = "0.2"
	maxOutputVersion = "0.2"
	// Technically Github allows 66536 characters, which we interpreted as 262144 bytes, but
	// we were still seeing 422 "Body is too long (maximum is 65536 characters)" errors so limit
	// more.
	GitHubMaxMessageSize = 200000 // bytes
)

type ReportInput struct {
	Metadata map[string]string
	Root     Root
}

// Load reads the file at the location p and the file body into a Root struct. Load naively
// validates that the Infracost JSON body is valid by checking the that the version attribute is within a supported range.
func Load(p string) (Root, error) {
	timer := metrics.GetTimer("output.load.duration", false).Start()
	defer timer.Stop()
	var out Root
	_, err := os.Stat(p)
	if errors.Is(err, os.ErrNotExist) {
		return out, errors.New("Infracost JSON file does not exist, generate it by running the following command then try again:\ninfracost breakdown --path /code --format json --out-file infracost-base.json")
	}

	data, err := os.ReadFile(p)
	if err != nil {
		return out, fmt.Errorf("error reading Infracost JSON file %w", err)
	}

	err = json.Unmarshal(data, &out)
	if err != nil {
		return out, fmt.Errorf("invalid Infracost JSON file %w, generate it by running the following command then try again:\ninfracost breakdown --path /code --format json --out-file infracost-base.json", err)
	}

	if !checkOutputVersion(out.Version) {
		return out, fmt.Errorf("invalid Infracost JSON file version. Supported versions are %s ≤ x ≤ %s", minOutputVersion, maxOutputVersion)
	}

	for i, p := range out.Projects {
		if p.Metadata == nil {
			p.Metadata = &schema.ProjectMetadata{}
			out.Projects[i] = p
		}

		if p.PastBreakdown == nil {
			p.PastBreakdown = &Breakdown{}
			out.Projects[i] = p
		}
	}

	return out, nil
}

func LoadPaths(paths []string) ([]ReportInput, error) {
	inputFiles := []string{}

	for _, path := range paths {
		// To make things easier in GitHub actions and other CI environments, we allow path to be a json array, e.g.:
		// --path='["/path/one", "/path/two"]'
		var nestedPaths []string
		err := json.Unmarshal([]byte(path), &nestedPaths)
		if err != nil {
			// This is not a json string so there must be no nested paths
			nestedPaths = []string{path}
		}

		for _, p := range nestedPaths {
			expanded, err := homedir.Expand(p)
			if err != nil {
				return nil, errors.Wrap(err, "Failed to expand path")
			}

			matches, _ := filepath.Glob(expanded)
			if len(matches) > 0 {
				inputFiles = append(inputFiles, matches...)
			} else {
				inputFiles = append(inputFiles, p)
			}
		}
	}

	inputs := make([]ReportInput, 0, len(inputFiles))

	for _, f := range inputFiles {
		r, err := Load(f)
		if err != nil {
			return nil, fmt.Errorf("could not load input file %s err: %w", f, err)
		}

		inputs = append(inputs, ReportInput{
			Metadata: map[string]string{
				"filename": f,
			},
			Root: r,
		})
	}

	return inputs, nil
}

// CompareTo generates an output Root using another Root as the base snapshot.
// Each project in current Root will have all past resources overwritten with the matching projects
// in the prior Root. If we can't find a matching project then we assume that the project
// has been newly created and will show a 100% increase in the output Root.
func CompareTo(c *config.Config, current, prior Root) (Root, error) {
	compareTimer := metrics.GetTimer("output.compare.duration", false).Start()
	defer compareTimer.Stop()
	currentProjectLabels := make(map[string]bool, len(current.Projects))
	for _, p := range current.Projects {
		currentProjectLabels[p.LabelWithMetadata()] = true
	}
	priorProjects := make(map[string]*schema.Project)
	for _, p := range prior.Projects {
		if _, ok := priorProjects[p.LabelWithMetadata()]; ok {
			return Root{}, fmt.Errorf("Invalid --compare-to Infracost JSON, found duplicate project name %s", p.LabelWithMetadata())
		}

		priorProjects[p.LabelWithMetadata()] = p.ToSchemaProject()
	}

	var schemaProjects schema.Projects
	for _, p := range current.Projects {
		scp := p.ToSchemaProject()
		scp.Diff = scp.Resources
		scp.PastResources = nil
		scp.Metadata.PastPolicySha = ""
		scp.HasDiff = true
		scp.Metadata.CurrentErrors = scp.Metadata.Errors

		metadata := p.LabelWithMetadata()
		if v, ok := priorProjects[metadata]; ok {
			if !p.Metadata.HasErrors() && !v.Metadata.HasErrors() {
				scp.PastResources = v.Resources
				scp.Metadata.PastPolicySha = v.Metadata.PolicySha
				scp.Diff = schema.CalculateDiff(scp.PastResources, scp.Resources)
			}

			if !p.Metadata.HasErrors() && !v.Metadata.IsEmptyProjectError() && v.Metadata.HasErrors() {
				// the prior project has errors, but the current one does not
				// The prior errors will be copied over to the current, but we
				// also need to remove the current project costs
				scp.Resources = nil
				scp.Diff = nil
				scp.HasDiff = false
			}

			for _, pastE := range v.Metadata.Errors {
				if schema.IsEmptyPathTypeError(pastE) {
					// If the error is a path type error we want to remove it from the metadata as
					// this is normally indicative of a project that has been added. Thus the project
					// path only appears in the current branch and so the baseline error is safe to
					// be ignored.
					continue
				}

				scp.Metadata.PastErrors = append(scp.Metadata.PastErrors, pastE)

				pastE.Message = "Diff baseline error: " + pastE.Message
				scp.Metadata.Errors = append(scp.Metadata.Errors, pastE)
			}

			delete(priorProjects, metadata)
		} else if children := findChildrenOfErroredProject(p, priorProjects); len(children) > 0 {
			for _, child := range children {
				if _, ok := currentProjectLabels[child]; ok {
					// this child has a match in the current projects so it should not be deleted
					continue
				}
				delete(priorProjects, child)
			}
		}

		schemaProjects = append(schemaProjects, scp)
	}

	for _, scp := range priorProjects {
		scp.PastResources = scp.Resources
		scp.Resources = nil
		scp.Metadata.PolicySha = ""
		scp.HasDiff = true
		scp.Diff = schema.CalculateDiff(scp.PastResources, scp.Resources)

		schemaProjects = append(schemaProjects, scp)
	}

	sort.Sort(schemaProjects)

	out, err := ToOutputFormat(c, schemaProjects)
	if err != nil {
		return out, err
	}

	// preserve the summary from the original run
	currentProjects := make(map[string]Project)
	for _, p := range current.Projects {
		currentProjects[p.LabelWithMetadata()] = p
	}
	for i := range out.Projects {
		if v, ok := currentProjects[out.Projects[i].LabelWithMetadata()]; ok {
			out.Projects[i].Summary = v.Summary
			out.Projects[i].fullSummary = v.fullSummary
		}
	}

	out.Summary = current.Summary
	out.FullSummary = current.FullSummary
	out.Currency = current.Currency
	return out, nil
}

// findChildrenOfErroredProject finds all the projects which are a child path of the errored Project p.
// This is done as sometimes errored Terragrunt evaluation returns the master path of the project
// rather than the actual Terragrunt projects. This happens when Terragrunt is unable to build
// a "stack" configuration (the Terragrunt project tree) because of a parsing error.
//
// For example if we have the following tree:
// .
// └── infra/
//
//	├── dev/
//	│   └── terragrunt.hcl
//	└── prod/
//	    └── terragrunt.hcl
//
// A valid `breakdown --path infra` will return projects `infra/dev` and `infra/prod`.
// However, in an errored "stack" state Terragrunt will return a single project at `infra`.
// We need to find all the child projects of the parent so that we can properly exclude them from the output.
// Otherwise, these projects will show as "removed" and have an invalid cost decrease.
//
// findChildrenOfErroredProject returns a list of strings that represent the paths of the projects
// to remove from an output list.
func findChildrenOfErroredProject(p Project, projects map[string]*schema.Project) []string {
	if !p.Metadata.HasErrors() {
		return nil
	}

	metadata := p.LabelWithMetadata()
	var children []string
	for key, project := range projects {
		if strings.HasPrefix(key, metadata) && !project.Metadata.HasErrors() {
			children = append(children, key)
		}
	}

	return children
}

func Combine(inputs []ReportInput) (Root, error) {
	var combined Root

	var lastestGeneratedAt time.Time
	var totalHourlyCost *decimal.Decimal
	var totalMonthlyCost *decimal.Decimal
	var totalMonthlyUsageCost *decimal.Decimal
	var pastTotalHourlyCost *decimal.Decimal
	var pastTotalMonthlyCost *decimal.Decimal
	var pastTotalMonthlyUsageCost *decimal.Decimal
	var diffTotalHourlyCost *decimal.Decimal
	var diffTotalMonthlyCost *decimal.Decimal
	var diffTotalMonthlyUsageCost *decimal.Decimal

	projects := make([]Project, 0)
	summaries := make([]*Summary, 0, len(inputs))
	currency := ""

	var metadata Metadata
	var invalidMetadata bool
	builder := strings.Builder{}
	for i, input := range inputs {
		var err error
		currency, err = checkCurrency(currency, input.Root.Currency)
		if err != nil {
			return combined, err
		}

		projects = append(projects, input.Root.Projects...)

		summaries = append(summaries, input.Root.Summary)

		if input.Root.TimeGenerated.After(lastestGeneratedAt) {
			lastestGeneratedAt = input.Root.TimeGenerated
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
		if input.Root.TotalMonthlyUsageCost != nil {
			if totalMonthlyUsageCost == nil {
				totalMonthlyUsageCost = decimalPtr(decimal.Zero)
			}

			totalMonthlyUsageCost = decimalPtr(totalMonthlyUsageCost.Add(*input.Root.TotalMonthlyUsageCost))
		}
		if input.Root.PastTotalHourlyCost != nil {
			if pastTotalHourlyCost == nil {
				pastTotalHourlyCost = decimalPtr(decimal.Zero)
			}

			pastTotalHourlyCost = decimalPtr(pastTotalHourlyCost.Add(*input.Root.PastTotalHourlyCost))
		}
		if input.Root.PastTotalMonthlyCost != nil {
			if pastTotalMonthlyCost == nil {
				pastTotalMonthlyCost = decimalPtr(decimal.Zero)
			}

			pastTotalMonthlyCost = decimalPtr(pastTotalMonthlyCost.Add(*input.Root.PastTotalMonthlyCost))
		}
		if input.Root.PastTotalMonthlyUsageCost != nil {
			if pastTotalMonthlyUsageCost == nil {
				pastTotalMonthlyUsageCost = decimalPtr(decimal.Zero)
			}

			pastTotalMonthlyUsageCost = decimalPtr(pastTotalMonthlyUsageCost.Add(*input.Root.PastTotalMonthlyUsageCost))
		}
		if input.Root.DiffTotalMonthlyCost != nil {
			if diffTotalMonthlyCost == nil {
				diffTotalMonthlyCost = decimalPtr(decimal.Zero)
			}

			diffTotalMonthlyCost = decimalPtr(diffTotalMonthlyCost.Add(*input.Root.DiffTotalMonthlyCost))
		}
		if input.Root.DiffTotalMonthlyUsageCost != nil {
			if diffTotalMonthlyUsageCost == nil {
				diffTotalMonthlyUsageCost = decimalPtr(decimal.Zero)
			}

			diffTotalMonthlyUsageCost = decimalPtr(diffTotalMonthlyUsageCost.Add(*input.Root.DiffTotalMonthlyUsageCost))
		}
		if input.Root.DiffTotalHourlyCost != nil {
			if diffTotalHourlyCost == nil {
				diffTotalHourlyCost = decimalPtr(decimal.Zero)
			}

			diffTotalHourlyCost = decimalPtr(diffTotalHourlyCost.Add(*input.Root.DiffTotalHourlyCost))
		}

		if i != 0 && metadata.VCSRepositoryURL != input.Root.Metadata.VCSRepositoryURL {
			invalidMetadata = true
		}

		metadata = input.Root.Metadata
		builder.WriteString(fmt.Sprintf("%q, ", input.Root.Metadata.VCSRepositoryURL))
	}

	combined.Version = outputVersion
	combined.Currency = currency
	combined.Projects = projects
	combined.TotalHourlyCost = totalHourlyCost
	combined.TotalMonthlyCost = totalMonthlyCost
	combined.TotalMonthlyUsageCost = totalMonthlyUsageCost
	combined.PastTotalHourlyCost = pastTotalHourlyCost
	combined.PastTotalMonthlyCost = pastTotalMonthlyCost
	combined.PastTotalMonthlyUsageCost = pastTotalMonthlyUsageCost
	combined.DiffTotalHourlyCost = diffTotalHourlyCost
	combined.DiffTotalMonthlyCost = diffTotalMonthlyCost
	combined.DiffTotalMonthlyUsageCost = diffTotalMonthlyUsageCost
	combined.TimeGenerated = lastestGeneratedAt
	combined.Summary = MergeSummaries(summaries)
	combined.Metadata = metadata
	if len(inputs) > 0 {
		combined.CloudURL = inputs[len(inputs)-1].Root.CloudURL
	}

	if invalidMetadata {
		return combined, clierror.NewWarningF(
			"combining Infracost JSON for different VCS repositories %s. Using %s as the top-level repository in the outputted JSON",
			strings.TrimRight(builder.String(), ", "),
			metadata.VCSRepositoryURL,
		)
	}

	return combined, nil
}

func checkCurrency(inputCurrency, fileCurrency string) (string, error) {
	if fileCurrency == "" {
		fileCurrency = "USD" // default to USD
	}

	if inputCurrency == "" {
		// this must be the first file, save the input currency
		inputCurrency = fileCurrency
	}

	if inputCurrency != fileCurrency {
		return "", fmt.Errorf("Invalid Infracost JSON file currency mismatch.  Can't combine %s and %s", inputCurrency, fileCurrency)
	}

	return inputCurrency, nil
}

func checkOutputVersion(v string) bool {
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	return semver.Compare(v, "v"+minOutputVersion) >= 0 && semver.Compare(v, "v"+maxOutputVersion) <= 0
}

// FormatOutput returns Root r as the format specified. The default format is a table output.
func FormatOutput(format string, r Root, opts Options) ([]byte, error) {
	var b []byte
	var err error

	if opts.CurrencyFormat != "" {
		addCurrencyFormat(opts.CurrencyFormat)
	}

	switch format {
	case "json":
		b, err = ToJSON(r, opts)
	case "html":
		b, err = ToHTML(r, opts)
	case "diff":
		b, err = ToDiff(r, opts)
	case "github-comment":
		out, error := ToMarkdown(r, opts, MarkdownOptions{MaxMessageSize: GitHubMaxMessageSize})
		b, err = out.Msg, error
	case "gitlab-comment", "azure-repos-comment":
		out, error := ToMarkdown(r, opts, MarkdownOptions{})
		b, err = out.Msg, error
	case "bitbucket-comment":
		out, error := ToMarkdown(r, opts, MarkdownOptions{BasicSyntax: true})
		b, err = out.Msg, error
	case "bitbucket-comment-summary":
		out, error := ToMarkdown(r, opts, MarkdownOptions{BasicSyntax: true, OmitDetails: true})
		b, err = out.Msg, error
	case "slack-message":
		b, err = ToSlackMessage(r, opts)
	default:
		b, err = ToTable(r, opts)
	}

	if err != nil {
		return nil, fmt.Errorf("error generating %s output %w", format, err)
	}

	return b, nil
}

func addCurrencyFormat(currencyFormat string) {
	rgx := regexp.MustCompile(`^(.{3}): (.*)1(,|\.)234(,|\.)?([0-9]*)?(.*)$`)
	m := rgx.FindStringSubmatch(currencyFormat)

	if len(m) == 0 {
		logging.Logger.Warn().Msgf("Invalid currency format: %s", currencyFormat)
		return
	}

	currency := m[1]

	graphemeWithSpace := m[2]
	grapheme := strings.TrimSpace(graphemeWithSpace)
	template := "$" + strings.Repeat(" ", len(graphemeWithSpace)-len(grapheme)) + "1"

	if graphemeWithSpace == "" {
		graphemeWithSpace = m[6]
		grapheme = strings.TrimSpace(graphemeWithSpace)
		template = "1" + strings.Repeat(" ", len(graphemeWithSpace)-len(grapheme)) + "$"
	}

	thousand := m[3]
	decimal := m[4]
	fraction := len(m[5])

	money.AddCurrency(currency, grapheme, template, decimal, thousand, fraction)
}
