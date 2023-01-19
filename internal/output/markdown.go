package output

import (
	"bufio"
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"
	"unicode/utf8"

	"github.com/Masterminds/sprig"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/ui"
)

func formatMarkdownCostChange(currency string, pastCost, cost *decimal.Decimal, skipPlusMinus bool) string {
	if pastCost != nil && pastCost.Equals(*cost) {
		return formatWholeDecimalCurrency(currency, decimal.Zero)
	}

	percentChange := formatPercentChange(pastCost, cost)
	if len(percentChange) > 0 {
		percentChange = " " + "(" + percentChange + ")"
	}

	plusMinus := "+"
	if skipPlusMinus {
		plusMinus = ""
	}

	// can't just use out.DiffTotalMonthlyCost because it isn't set if there is no past cost
	if pastCost != nil {
		d := cost.Sub(*pastCost)
		if skipPlusMinus {
			d = d.Abs()
			return formatCost(currency, &d) + percentChange
		}

		if d.LessThan(decimal.Zero) {
			plusMinus = ""
		}

		return plusMinus + formatCost(currency, &d) + percentChange
	}

	return plusMinus + formatCost(currency, cost) + percentChange
}

func formatCostChangeSentence(currency string, pastCost, cost *decimal.Decimal, useEmoji bool) string {
	up := "📈"
	down := "📉"

	if !useEmoji {
		up = "↑"
		down = "↓"
	}

	if pastCost != nil {
		if pastCost.Equals(*cost) {
			return "monthly cost will not change"
		} else if pastCost.GreaterThan(*cost) {
			return "monthly cost will decrease by " + formatMarkdownCostChange(currency, pastCost, cost, true) + " " + down
		}
	}
	return "monthly cost will increase by " + formatMarkdownCostChange(currency, pastCost, cost, true) + " " + up
}

// ProjectRow is used to display the summary table of projects in the markdown output.
type ProjectRow struct {
	Name                 string
	ModulePath           string
	Workspace            string
	PastTotalMonthlyCost *decimal.Decimal
	TotalMonthlyCost     *decimal.Decimal
	DiffTotalMonthlyCost *decimal.Decimal
	HasErrors            bool
	VCSCodeChanged       bool
	HasDiff              bool
}

func calculateProjectRows(projects []Project) []ProjectRow {
	rows := make([]ProjectRow, 0, len(projects))

	pm := make(map[string][]Project)

	for _, p := range projects {
		pm[p.Name] = append(pm[p.Name], p)
	}


	for name, projects := range pm {

		var pastTotalMonthlyCost, totalMonthlyCost, diffTotalMonthlyCost *decimal.Decimal
		hasError, hasDiff, vcsCodeChanged := false, false, false

		for _, p := range projects {
			if p.Metadata.HasErrors() {
				hasError = true
			}

			if p.Diff != nil && len(p.Diff.Resources) > 0 {
				hasDiff = true
			}

			if p.Metadata.VCSCodeChanged != nil && *p.Metadata.VCSCodeChanged {
				vcsCodeChanged = true
			}

			if p.PastBreakdown != nil && p.PastBreakdown.TotalMonthlyCost != nil {
				if pastTotalMonthlyCost == nil {
					pastTotalMonthlyCost = decimalPtr(decimal.NewFromInt(0))
				}
				pastTotalMonthlyCost = decimalPtr(pastTotalMonthlyCost.Add(*p.PastBreakdown.TotalMonthlyCost))
			}

			if p.Breakdown != nil && p.Breakdown.TotalMonthlyCost != nil {
				if totalMonthlyCost == nil {
					totalMonthlyCost = decimalPtr(decimal.NewFromInt(0))
				}
				totalMonthlyCost = decimalPtr(totalMonthlyCost.Add(*p.Breakdown.TotalMonthlyCost))
			}

			if p.Diff != nil && p.Diff.TotalMonthlyCost != nil {
				if diffTotalMonthlyCost == nil {
					diffTotalMonthlyCost = decimalPtr(decimal.NewFromInt(0))
				}
				diffTotalMonthlyCost = decimalPtr(diffTotalMonthlyCost.Add(*p.Diff.TotalMonthlyCost))
			}
		}

		modulePath, workspace := "", ""
		if len(projects) == 1 {
			modulePath = projects[0].Metadata.TerraformModulePath
			workspace = projects[0].Metadata.WorkspaceLabel()
		}

		rows = append(rows, ProjectRow{
			Name:                 name,
			ModulePath:           modulePath,
			Workspace:            workspace,
			PastTotalMonthlyCost: pastTotalMonthlyCost,
			TotalMonthlyCost:     totalMonthlyCost,
			DiffTotalMonthlyCost: diffTotalMonthlyCost,
			HasErrors:            hasError,
			VCSCodeChanged:       vcsCodeChanged,
			HasDiff:              hasDiff,
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Name < rows[j].Name
	})

	return rows
}

func calculateMetadataToDisplay(rows []ProjectRow) (hasModulePath bool, hasWorkspace bool) {
	// we only want to show metadata fields if they can help distinguish projects with the same name

	srows := make([]ProjectRow, 0, len(rows))
	srows = append(srows, rows...)

	sort.Slice(srows, func(i, j int) bool {
		if srows[i].Name != srows[j].Name {
			return srows[i].Name < srows[j].Name
		}

		if srows[i].ModulePath != srows[j].ModulePath {
			return srows[i].ModulePath < srows[j].ModulePath
		}

		return srows[i].Workspace < srows[j].Workspace
	})

	// check if any projects that have the same name have different path or workspace
	for i, p := range srows {
		if i > 0 { // we compare vs the previous item, so skip index 0
			prev := srows[i-1]
			if p.Name == prev.Name {
				if p.ModulePath != prev.ModulePath {
					hasModulePath = true
				}
				if p.Workspace != prev.Workspace {
					hasWorkspace = true
				}
			}
		}
	}

	return hasModulePath, hasWorkspace
}

// MarkdownCtx holds information that can be used and executed with a go template.
type MarkdownCtx struct {
	Root                         Root
	ProjectRows                  []ProjectRow
	SkippedProjectCount          int
	ErroredProjectCount          int
	SkippedUnchangedProjectCount int
	DiffOutput                   string
	Options                      Options
	MarkdownOptions              MarkdownOptions
}

// ProjectCounts returns a string that represents additional information about missing/errored projects.
func (m MarkdownCtx) ProjectCounts() string {
	out := ""
	if m.SkippedUnchangedProjectCount == 1 {
		out += "1 project has no code changes, "
	} else if m.SkippedUnchangedProjectCount > 0 {
		out += fmt.Sprintf("%d projects have no code changes, ", m.SkippedUnchangedProjectCount)
	} else if m.SkippedProjectCount == 1 {
		out += "1 project has no cost estimate changes, "
	} else if m.SkippedProjectCount > 0 {
		out += fmt.Sprintf("%d projects have no cost estimate changes, ", m.SkippedProjectCount)
	}

	if m.ErroredProjectCount == 1 {
		out += "1 project could not be evaluated"
	} else if m.ErroredProjectCount > 0 {
		out += fmt.Sprintf("%d projects could not be evaluated, ", m.ErroredProjectCount)
	}

	if out == "" {
		return out
	}

	return "\n" + strings.TrimSuffix(out, ", ") + "."
}

func ToMarkdown(out Root, opts Options, markdownOpts MarkdownOptions) ([]byte, error) {
	var diffMsg string

	if opts.diffMsg != "" {
		diffMsg = opts.diffMsg
	} else {
		diff, err := ToDiff(out, opts)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to generate diff")
		}

		diffMsg = ui.StripColor(string(diff))
	}

	projectRows := calculateProjectRows(out.Projects)
	hasModulePath, hasWorkspace := calculateMetadataToDisplay(projectRows)

	var buf bytes.Buffer
	bufw := bufio.NewWriter(&buf)

	tmpl := template.New("base")
	tmpl.Funcs(sprig.TxtFuncMap())
	tmpl.Funcs(template.FuncMap{
		"formatCost": func(d *decimal.Decimal) string {
			if d == nil || d.IsZero() {
				return formatWholeDecimalCurrency(out.Currency, decimal.Zero)
			}
			return formatCost(out.Currency, d)
		},
		"formatCostChange": func(pastCost, cost *decimal.Decimal) string {
			return formatMarkdownCostChange(out.Currency, pastCost, cost, false)
		},
		"formatCostChangeSentence": formatCostChangeSentence,
		"showProjectRow": func(p ProjectRow) bool {
			if p.HasErrors {
				return false
			}

			if opts.ShowOnlyChanges {
				// only return true if the project has code changes so the table can also show
				// project that have cost changes.
				if p.VCSCodeChanged {
					return true
				}
			}

			if opts.ShowAllProjects {
				return true
			}

			if !p.HasDiff {
				return false
			}

			return true // has diff
		},
		"metadataHeaders": func() []string {
			headers := []string{}
			if hasModulePath {
				headers = append(headers, "Module path")
			}
			if hasWorkspace {
				headers = append(headers, "Workspace")
			}
			return headers
		},
		"metadataFields": func(p ProjectRow) []string {
			fields := []string{}
			if hasModulePath {
				fields = append(fields, p.ModulePath)
			}
			if hasWorkspace {
				fields = append(fields, p.Workspace)
			}
			return fields
		},
		"metadataPlaceholders": func() []string {
			placeholders := []string{}
			if hasModulePath {
				placeholders = append(placeholders, "")
			}
			if hasWorkspace {
				placeholders = append(placeholders, "")
			}
			return placeholders
		},
		"truncateMiddle": truncateMiddle,
	})

	t := CommentMarkdownWithHTMLTemplate
	if markdownOpts.BasicSyntax {
		t = CommentMarkdownTemplate
	}
	tmpl, err := tmpl.Parse(t)
	if err != nil {
		return []byte{}, err
	}

	skippedProjectCount := 0
	for _, p := range projectRows {
		hasCodeChanges := opts.ShowOnlyChanges && p.VCSCodeChanged

		if !p.HasErrors && !p.HasDiff && !hasCodeChanges {
			skippedProjectCount++
		}
	}

	erroredProjectCount := 0
	for _, p := range projectRows {
		if p.HasErrors {
			erroredProjectCount++
		}
	}

	skippedUnchangedProjectCount := 0
	if opts.ShowOnlyChanges {
		for _, p := range projectRows {
			if !p.VCSCodeChanged {
				skippedUnchangedProjectCount++
			}
		}
	}

	err = tmpl.Execute(bufw, MarkdownCtx{
		out,
		projectRows,
		skippedProjectCount,
		erroredProjectCount,
		skippedUnchangedProjectCount,
		diffMsg,
		opts,
		markdownOpts})
	if err != nil {
		return []byte{}, err
	}

	bufw.Flush()
	msg := buf.Bytes()

	msgByteLength := len(msg)
	if markdownOpts.MaxMessageSize > 0 && msgByteLength > markdownOpts.MaxMessageSize {
		msgRuneLength := utf8.RuneCount(msg)
		// truncation relies on rune length
		q := float64(msgRuneLength) / float64(msgByteLength)
		truncateLength := msgRuneLength - int(q*float64(markdownOpts.MaxMessageSize))
		newLength := utf8.RuneCountInString(diffMsg) - truncateLength - 1000

		opts.diffMsg = truncateMiddle(diffMsg, newLength, "\n\n...(truncated due to message size limit)...\n\n")
		return ToMarkdown(out, opts, markdownOpts)
	}

	return msg, nil
}

func hasCodeChanges(options Options, project Project) bool {
	return options.ShowOnlyChanges && project.Metadata.VCSCodeChanged != nil && *project.Metadata.VCSCodeChanged
}
