package output

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"regexp"
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

func formatMarkdownCostChange(currency string, pastCost, cost *decimal.Decimal, skipPlusMinus, skipPercent, skipIfZero bool) string {
	if pastCost == nil && cost == nil {
		return "-"
	}

	if skipIfZero && (pastCost == nil || pastCost.IsZero()) && (cost == nil || cost.IsZero()) {
		return "-"
	}

	plusMinus := "+"
	if skipPlusMinus {
		plusMinus = ""
	}

	if pastCost != nil && cost != nil && pastCost.Equals(*cost) {
		return plusMinus + formatWholeDecimalCurrency(currency, decimal.Zero)
	}

	percentChange := ""
	if !skipPercent {
		percentChange = formatPercentChange(pastCost, cost)
		if len(percentChange) > 0 {
			percentChange = " " + "(" + percentChange + ")"
		}
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
		return "Monthly estimate increased by " + formatCost(currency, cost) + " " + up
	}

	diff := cost.Sub(*pastCost).Abs()
	change := formatCost(currency, &diff)

	if pastCost.Equals(*cost) {
		return "Monthly estimate generated"
	}

	if pastCost.GreaterThan(*cost) {
		return "Monthly estimate decreased by " + change + " " + down
	}

	return "Monthly estimate increased by " + change + " " + up
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
	RunQuotaMsg                  string
	UsageCostsMsg                string
	CostDetailsMsg               string
}

// MarkdownOutput holds the message converted to markdown with additional
// information about its length.
type MarkdownOutput struct {
	Msg             []byte
	RuneLen         int
	OriginalMsgSize int
}

func ToMarkdown(out Root, opts Options, markdownOpts MarkdownOptions) (MarkdownOutput, error) {
	var diffMsg string

	if opts.diffMsg != "" {
		diffMsg = opts.diffMsg
	} else {
		diff, err := ToDiff(out, opts)
		if err != nil {
			return MarkdownOutput{}, errors.Wrap(err, "Failed to generate diff")
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

	runQuotaMsg, exceeded := out.Projects.IsRunQuotaExceeded()
	if exceeded {
		filename = "run-quota-exceeded.tmpl"
	}

	skipUsageCostIfZero := !usageCostsEnabled(out)

	tmpl := template.New(filename)
	tmpl.Funcs(sprig.TxtFuncMap())
	tmpl.Funcs(template.FuncMap{
		"formatCost": func(d *decimal.Decimal) string {
			if d == nil || d.IsZero() {
				return formatWholeDecimalCurrency(out.Currency, decimal.Zero)
			}
			return formatCost(out.Currency, d)
		},
		"formatUsageCost": func(d *decimal.Decimal) string {
			return formatUsageCost(out, d)
		},
		"formatCostChange": func(pastCost, cost *decimal.Decimal) string {
			return formatMarkdownCostChange(out.Currency, pastCost, cost, false, false, false)
		},
		"formatCostChangeWithoutPercent": func(pastCost, cost *decimal.Decimal) string {
			return formatMarkdownCostChange(out.Currency, pastCost, cost, false, true, false)
		},
		"formatUsageCostChangeWithoutPercent": func(pastCost, cost *decimal.Decimal) string {
			return formatMarkdownCostChange(out.Currency, pastCost, cost, false, true, skipUsageCostIfZero)
		},
		"formatCostChangeSentence": formatCostChangeSentence,
		"showProject": func(p Project) bool {
			return showProject(p, opts, false)
		},
		"cloudURL": func() string {
			return out.CloudURL
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

			// we always want to show the output if there are unsupported resources. This is
			// because we want to show the unsupported resources in the output so that the
			// user can see why the cost changes are different from expectations.
			if out.HasUnsupportedResources() {
				return true
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
		return MarkdownOutput{}, err
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
		MarkdownOptions:              markdownOpts,
		RunQuotaMsg:                  runQuotaMsg,
		UsageCostsMsg:                usageCostsMessage(out, true),
		CostDetailsMsg:               costsDetailsMessage(out),
	})
	if err != nil {
		return MarkdownOutput{}, err
	}

	bufw.Flush()
	msg := buf.Bytes()

	msgByteLength := len(msg)
	msgRuneLength := utf8.RuneCount(msg)

	originalSize := msgRuneLength
	if opts.originalSize > 0 {
		originalSize = opts.originalSize
	}

	if markdownOpts.MaxMessageSize > 0 && msgByteLength > markdownOpts.MaxMessageSize {
		// Calculate how much we need to reduce the message size
		excessBytes := msgByteLength - markdownOpts.MaxMessageSize

		// Use the diff message's own rune-to-byte ratio for more accurate truncation
		diffMsgRunes := utf8.RuneCountInString(diffMsg)
		diffMsgBytes := len([]byte(diffMsg))

		var truncateRunes int
		if diffMsgBytes > 0 {
			diffRatio := float64(diffMsgRunes) / float64(diffMsgBytes)
			truncateRunes = int(float64(excessBytes) * diffRatio)
		} else {
			truncateRunes = excessBytes // fallback
		}

		newLength := diffMsgRunes - truncateRunes - 1000

		if newLength < 0 {
			// trimming diff msg is not enough, so we truncate the whole message
			truncated := truncateMiddle(string(msg), markdownOpts.MaxMessageSize-1000, "\n\n...(truncated due to message size limit)...\n\n")
			return MarkdownOutput{Msg: []byte(truncated), RuneLen: utf8.RuneCountInString(truncated), OriginalMsgSize: originalSize}, nil
		}

		opts.diffMsg = truncateMiddle(diffMsg, newLength, "\n\n...(truncated due to message size limit)...\n\n")
		opts.originalSize = originalSize
		return ToMarkdown(out, opts, markdownOpts)
	}

	return MarkdownOutput{Msg: msg, RuneLen: msgRuneLength, OriginalMsgSize: originalSize}, nil
}

func hasCodeChanges(options Options, project Project) bool {
	return options.ShowOnlyChanges && project.Metadata.VCSCodeChanged != nil && *project.Metadata.VCSCodeChanged
}

var (
	usageFileDocsURL             = "https://www.infracost.io/docs/features/usage_based_resources/#infracost-usageyml"
	usageDefaultsDocsURL         = "https://www.infracost.io/docs/features/usage_based_resources"
	usageDefaultsDashboardSuffix = "settings/usage-cost-defaults"
	cloudURLRegex                = regexp.MustCompile(`(https?://[^/]+/org/[^/]+/)`)
)

func usageCostsMessage(out Root, useLinks bool) string {
	if !out.Metadata.UsageApiEnabled {
		return handleUsageApiDisabledMessage(out, useLinks)
	}

	return handleUsageApiEnabledMessage(out, useLinks)
}

func cloudSettingsStr(out Root, useMarkdownLinks bool) string {
	if !useMarkdownLinks {
		return "Infracost Cloud settings"
	}
	usageDefaultsURL := usageDefaultsDocsURL
	match := cloudURLRegex.FindStringSubmatch(out.CloudURL)
	if len(match) > 0 {
		usageDefaultsURL = match[0] + usageDefaultsDashboardSuffix
	}
	return fmt.Sprintf("[Infracost Cloud settings](%s)", usageDefaultsURL)
}

func usageDocsStr(useMarkdownLinks bool) string {
	if !useMarkdownLinks {
		return "docs"
	}

	return fmt.Sprintf("[docs](%s)", usageFileDocsURL)
}

func handleUsageApiEnabledMessage(out Root, useMarkdownLinks bool) string {
	if out.Metadata.UsageFilePath != "" || out.Metadata.ConfigFileHasUsageFile {
		return fmt.Sprintf("*Usage costs were estimated by merging infracost-usage.yml and %s.", cloudSettingsStr(out, useMarkdownLinks))
	}

	return fmt.Sprintf("*Usage costs were estimated using %s, see %s for other options.", cloudSettingsStr(out, useMarkdownLinks), usageDocsStr(useMarkdownLinks))
}

func handleUsageApiDisabledMessage(out Root, useMarkdownLinks bool) string {
	if out.Metadata.UsageFilePath != "" || out.Metadata.ConfigFileHasUsageFile {
		return fmt.Sprintf("*Usage costs were estimated using infracost-usage.yml, see %s for other options.", usageDocsStr(useMarkdownLinks))
	}

	return fmt.Sprintf("*Usage costs can be estimated by updating %s, see %s for other options.", cloudSettingsStr(out, useMarkdownLinks), usageDocsStr(useMarkdownLinks))
}

func costsDetailsMessage(out Root) string {
	var msgs []string

	if out.Summary != nil && out.Summary.TotalUnsupportedResources != nil && *out.Summary.TotalUnsupportedResources > 0 {
		msgs = append(msgs, "unsupported resources")
	}

	for _, p := range out.Projects {
		if len(p.Metadata.Errors) > 0 {
			msgs = append(msgs, "skipped projects due to errors")
			break
		}
	}

	if len(msgs) == 0 {
		return ""
	}

	return fmt.Sprintf("(includes details of %s)", strings.Join(msgs, " and "))
}
