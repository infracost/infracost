package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"golang.org/x/mod/semver"

	"github.com/infracost/infracost/internal/schema"
)

var (
	minOutputVersion = "0.2"
	maxOutputVersion = "0.2"
)

type ReportInput struct {
	Metadata map[string]string
	Root     Root
}

// Load reads the file at the location p and the file body into a Root struct. Load naively
// validates that the Infracost JSON body is valid by checking the that the version attribute is within a supported range.
func Load(p string) (Root, error) {
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
func CompareTo(current, prior Root) (Root, error) {
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
		scp.HasDiff = true

		if v, ok := priorProjects[p.LabelWithMetadata()]; ok {
			scp.PastResources = v.Resources
			scp.Diff = schema.CalculateDiff(scp.PastResources, scp.Resources)
			delete(priorProjects, p.LabelWithMetadata())
		}

		schemaProjects = append(schemaProjects, scp)
	}

	for _, scp := range priorProjects {
		scp.PastResources = scp.Resources
		scp.Resources = nil
		scp.HasDiff = true
		scp.Diff = schema.CalculateDiff(scp.PastResources, scp.Resources)

		schemaProjects = append(schemaProjects, scp)
	}

	sort.Sort(schemaProjects)

	out, err := ToOutputFormat(schemaProjects)
	if err != nil {
		return out, err
	}

	out.Currency = current.Currency
	return out, nil
}

func Combine(inputs []ReportInput) (Root, error) {
	var combined Root

	var totalHourlyCost *decimal.Decimal
	var totalMonthlyCost *decimal.Decimal
	var pastTotalHourlyCost *decimal.Decimal
	var pastTotalMonthlyCost *decimal.Decimal
	var diffTotalHourlyCost *decimal.Decimal
	var diffTotalMonthlyCost *decimal.Decimal

	projects := make([]Project, 0)
	summaries := make([]*Summary, 0, len(inputs))
	currency := ""

	for _, input := range inputs {
		var err error
		currency, err = checkCurrency(currency, input.Root.Currency)
		if err != nil {
			return combined, err
		}

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
		if input.Root.DiffTotalMonthlyCost != nil {
			if diffTotalMonthlyCost == nil {
				diffTotalMonthlyCost = decimalPtr(decimal.Zero)
			}

			diffTotalMonthlyCost = decimalPtr(diffTotalMonthlyCost.Add(*input.Root.DiffTotalMonthlyCost))
		}

		if input.Root.DiffTotalHourlyCost != nil {
			if diffTotalHourlyCost == nil {
				diffTotalHourlyCost = decimalPtr(decimal.Zero)
			}

			diffTotalHourlyCost = decimalPtr(diffTotalHourlyCost.Add(*input.Root.DiffTotalHourlyCost))
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
	combined.TimeGenerated = time.Now().UTC()
	combined.Summary = MergeSummaries(summaries)

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

	switch format {
	case "json":
		b, err = ToJSON(r, opts)
	case "html":
		b, err = ToHTML(r, opts)
	case "diff":
		b, err = ToDiff(r, opts)
	case "github-comment", "gitlab-comment", "azure-repos-comment":
		b, err = ToMarkdown(r, opts, MarkdownOptions{})
	case "bitbucket-comment":
		b, err = ToMarkdown(r, opts, MarkdownOptions{BasicSyntax: true})
	case "bitbucket-comment-summary":
		b, err = ToMarkdown(r, opts, MarkdownOptions{BasicSyntax: true, OmitDetails: true})
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
