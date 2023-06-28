package output

import (
	"bufio"
	"bytes"
	"embed"
	"sort"
	"strings"
	"text/template"
	"unicode/utf8"

	"github.com/Masterminds/sprig"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/ui"
)

//go:embed templates/*
var templatesFS embed.FS

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

	if pastCost == nil {
		return "monthly cost will increase by " + formatCost(currency, cost) + " " + up
	}

	diff := cost.Sub(*pastCost).Abs()
	change := formatCost(currency, &diff)

	if pastCost.Equals(*cost) {
		return "monthly cost will not change"
	}

	if pastCost.GreaterThan(*cost) {
		return "monthly cost will decrease by " + change + " " + down
	}

	return "monthly cost will increase by " + change + " " + up
}

func calculateMetadataToDisplay(projects []Project) (hasModulePath bool, hasWorkspace bool) {
	sort.Slice(projects, func(i, j int) bool {
		if projects[i].Name != projects[j].Name {
			return projects[i].Name < projects[j].Name
		}

		if projects[i].Metadata.TerraformModulePath != projects[j].Metadata.TerraformModulePath {
			return projects[i].Metadata.TerraformModulePath < projects[j].Metadata.TerraformModulePath
		}

		return projects[i].Metadata.WorkspaceLabel() < projects[j].Metadata.WorkspaceLabel()
	})

	// check if any projects that have the same name have different path or workspace
	for i, p := range projects {
		if i > 0 { // we compare vs the previous item, so skip index 0
			prev := projects[i-1]
			if p.Name == prev.Name {
				if p.Metadata.TerraformModulePath != prev.Metadata.TerraformModulePath {
					hasModulePath = true
				}
				if p.Metadata.WorkspaceLabel() != prev.Metadata.WorkspaceLabel() {
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
	SkippedProjectCount          int
	ErroredProjectCount          int
	SkippedUnchangedProjectCount int
	DiffOutput                   string
	Options                      Options
	MarkdownOptions              MarkdownOptions
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

	hasModulePath, hasWorkspace := calculateMetadataToDisplay(out.Projects)

	var buf bytes.Buffer
	bufw := bufio.NewWriter(&buf)

	filename := "markdown-html.tmpl"
	if markdownOpts.BasicSyntax {
		filename = "markdown.tmpl"
	}

	tmpl := template.New(filename)
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
		"showProject": func(p Project) bool {
			return showProject(p, opts, false)
		},
		"displaySub": func() bool {
			if out.CloudURL != "" {
				return true
			}

			return markdownOpts.WillUpdate || markdownOpts.WillReplace
		},
		"displayTable": func() bool {
			var valid Projects
			for _, project := range out.Projects {
				if showProject(project, opts, false) {
					valid = append(valid, project)
				}
			}

			return len(valid) > 0
		},
		"displayOutput": func() bool {
			if markdownOpts.OmitDetails {
				return false
			}

			var valid Projects
			for _, project := range out.Projects {
				if showProject(project, opts, true) {
					valid = append(valid, project)
				}
			}

			return len(valid) > 0
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
		"metadataFields": func(p Project) []string {
			fields := []string{}
			if hasModulePath {
				fields = append(fields, p.Metadata.TerraformModulePath)
			}
			if hasWorkspace {
				fields = append(fields, p.Metadata.WorkspaceLabel())
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
		"stringsJoin":    strings.Join,
		"truncateMiddle": truncateMiddle,
	})
	_, err := tmpl.ParseFS(templatesFS, "templates/"+filename)
	if err != nil {
		return []byte{}, err
	}

	skippedProjectCount := 0
	for _, p := range out.Projects {
		if p.Metadata.HasErrors() {
			continue
		}

		if (p.Diff == nil || len(p.Diff.Resources) == 0) && !hasCodeChanges(opts, p) {
			skippedProjectCount++
		}
	}

	erroredProjectCount := 0
	for _, p := range out.Projects {
		if p.Metadata.HasErrors() {
			erroredProjectCount++
		}
	}

	skippedUnchangedProjectCount := 0
	if opts.ShowOnlyChanges {
		for _, p := range out.Projects {
			if !hasCodeChanges(opts, p) {
				skippedUnchangedProjectCount++
			}
		}
	}

	err = tmpl.Execute(bufw, MarkdownCtx{
		Root:                         out,
		SkippedProjectCount:          skippedProjectCount,
		ErroredProjectCount:          erroredProjectCount,
		SkippedUnchangedProjectCount: skippedUnchangedProjectCount,
		DiffOutput:                   diffMsg,
		Options:                      opts,
		MarkdownOptions:              markdownOpts})
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
