package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"golang.org/x/mod/semver"
)

var minOutputVersion = "0.2"
var maxOutputVersion = "0.2"

type ReportInput struct {
	Metadata map[string]string
	Root     Root
}

func Load(data []byte) (Root, error) {
	var out Root
	err := json.Unmarshal(data, &out)
	return out, err
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
		data, err := os.ReadFile(f)
		if err != nil {
			return nil, errors.Wrap(err, "Error reading JSON file")
		}

		j, err := Load(data)
		if err != nil {
			return nil, errors.Wrap(err, "Error parsing JSON file")
		}

		if !checkOutputVersion(j.Version) {
			return nil, fmt.Errorf("Invalid Infracost JSON file version. Supported versions are %s ≤ x ≤ %s", minOutputVersion, maxOutputVersion)
		}

		inputs = append(inputs, ReportInput{
			Metadata: map[string]string{
				"filename": f,
			},
			Root: j,
		})
	}

	return inputs, nil
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
	combined.TimeGenerated = time.Now()
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
