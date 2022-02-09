package output

import (
	"bufio"
	"bytes"
	"text/template"

	"github.com/infracost/infracost/internal/ui"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	"github.com/Masterminds/sprig"
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

func formatCostChangeSentence(currency string, pastCost, cost *decimal.Decimal) string {
	if pastCost != nil {
		if pastCost.Equals(*cost) {
			return "monthly cost will not change"
		} else if pastCost.GreaterThan(*cost) {
			return "monthly cost will decrease by " + formatMarkdownCostChange(currency, pastCost, cost, true) + " 📉"
		}
	}
	return "monthly cost will increase by " + formatMarkdownCostChange(currency, pastCost, cost, true) + " 📈"
}

func ToMarkdown(out Root, opts Options, markdownOpts MarkdownOptions) ([]byte, error) {
	diff, err := ToDiff(out, opts)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to generate diff")
	}

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
		"hasDiff": func(p Project) bool {
			if p.Diff == nil || len(p.Diff.Resources) == 0 {
				return false
			}
			return true
		},
		"projectLabel": func(p Project) string {
			return p.Label(opts.DashboardEnabled)
		},
		"truncateMiddle": truncateMiddle,
	})

	t := CommentMarkdownWithHTMLTemplate
	if markdownOpts.BasicSyntax {
		t = CommentMarkdownTemplate
	}
	tmpl, err = tmpl.Parse(t)
	if err != nil {
		return []byte{}, err
	}

	skippedProjectCount := 0
	for _, p := range out.Projects {
		if p.Diff == nil || len(p.Diff.Resources) == 0 {
			skippedProjectCount++
		}
	}

	err = tmpl.Execute(bufw, struct {
		Root                Root
		SkippedProjectCount int
		DiffOutput          string
		Options             Options
		MarkdownOptions     MarkdownOptions
	}{
		out,
		skippedProjectCount,
		ui.StripColor(string(diff)),
		opts,
		markdownOpts})
	if err != nil {
		return []byte{}, err
	}

	bufw.Flush()
	return buf.Bytes(), nil
}
