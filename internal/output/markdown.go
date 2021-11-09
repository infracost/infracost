package output

import (
	"bufio"
	"bytes"
	"github.com/infracost/infracost/internal/ui"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
)

func ToMarkdown(out Root, opts Options) ([]byte, error) {
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
				return "-"
			}
			return formatCost(out.Currency, d)
		},
		"formatCostChange": func(pastCost, cost *decimal.Decimal) string {
			if pastCost != nil && pastCost.Equals(*cost) {
				return "-"
			}

			percentChange := formatPercentChange(pastCost, cost)
			if len(percentChange) > 0 {
				percentChange = " " + "(" + percentChange + ")"
			}

			// can't just use out.DiffTotalMonthlyCost because it isn't set if there is no past cost
			if pastCost != nil {
				d := cost.Sub(*pastCost)
				return formatCost(out.Currency, &d) + percentChange
			}
			return formatCost(out.Currency, cost) + percentChange
		},
		"increasing": func(pastCost, cost *decimal.Decimal) bool { return pastCost == nil || cost.GreaterThan(*pastCost) },
		"hasDiff": func(p Project) bool {
			if p.Diff == nil || len(p.Diff.Resources) == 0 {
				return false
			}
			return true
		},
		"projectLabel": func(p Project) string {
			return p.Label(opts.DashboardEnabled)
		},
		"truncateProjectName": func(name string) string {
			maxProjectNameLength := 64
			// Truncate the middle of the name if it's too long
			if len(name) > maxProjectNameLength {
				return name[0:(maxProjectNameLength/2)-1] + "..." + name[len(name)-(maxProjectNameLength/2+3):]
			}
			return name
		},
	})
	tmpl, err = tmpl.Parse(CommentMarkdownTemplate)
	if err != nil {
		return []byte{}, err
	}

	skippedProjects := make([]string, 0)
	for _, p := range out.Projects {
		if p.Diff == nil || len(p.Diff.Resources) == 0 {
			skippedProjects = append(skippedProjects, p.Name)
		}
	}

	err = tmpl.Execute(bufw, struct {
		Root            Root
		SkippedProjects string
		DiffOutput      string
		WillUpdate      bool
		Options         Options
	}{
		out,
		strings.Join(skippedProjects, ", "),
		ui.StripColor(string(diff)),
		true,
		opts})
	if err != nil {
		return []byte{}, err
	}

	bufw.Flush()
	return buf.Bytes(), nil
}
