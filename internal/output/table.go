package output

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/ui"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func ToTable(out Root, opts Options) ([]byte, error) {
	var tableLen int

	s := ""

	hasNilCosts := false

	// Don't show the project total if there's only one project result
	// since we will show the overall total anyway
	includeProjectTotals := len(out.Projects) != 1

	for i, project := range out.Projects {
		if project.Breakdown == nil {
			continue
		}

		if i != 0 {
			s += "----------------------------------\n"
		}

		s += fmt.Sprintf("%s %s\n\n",
			ui.BoldString("Project:"),
			project.Label(),
		)

		if breakdownHasNilCosts(*project.Breakdown) {
			hasNilCosts = true
		}

		tableOut := tableForBreakdown(*project.Breakdown, opts.Fields, includeProjectTotals)

		// Get the last table length so we can align the overall total with it
		if i == len(out.Projects)-1 {
			tableLen = len(ui.StripColor(strings.SplitN(tableOut, "\n", 2)[0]))
		}

		s += tableOut

		s += "\n"

		if i != len(out.Projects)-1 {
			s += "\n"
		}
	}

	if includeProjectTotals {
		s += "\n"
	}

	totalOut := formatCost2DP(out.TotalMonthlyCost)

	s += fmt.Sprintf("%s%s",
		ui.BoldString(" OVERALL TOTAL"),
		fmt.Sprintf("%*s ", tableLen-15, totalOut), // pad based on the last line length
	)

	unsupportedMsg := out.unsupportedResourcesMessage(opts.ShowSkipped)

	if hasNilCosts || unsupportedMsg != "" {
		s += "\n----------------------------------"
	}

	if hasNilCosts {
		s += fmt.Sprintf("\nTo estimate usage-based resources use --usage-file, see %s",
			ui.LinkString("https://infracost.io/usage-file"),
		)

		if unsupportedMsg != "" {
			s += "\n"
		}
	}

	if unsupportedMsg != "" {
		s += "\n" + unsupportedMsg
	}

	return []byte(s), nil
}

func tableForBreakdown(breakdown Breakdown, fields []string, includeTotal bool) string {
	t := table.NewWriter()
	t.Style().Options.DrawBorder = false
	t.Style().Options.SeparateColumns = false
	t.Style().Options.SeparateRows = false
	t.Style().Options.SeparateHeader = false
	t.Style().Format.Header = text.FormatDefault

	var columns []table.ColumnConfig
	var headers table.Row
	headers = append(headers,
		ui.UnderlineString("Name"),
	)

	i := 1

	columns = append(columns, table.ColumnConfig{
		Number:      i,
		Align:       text.AlignLeft,
		AlignHeader: text.AlignLeft,
	})
	i++

	if contains(fields, "price") {
		headers = append(headers, ui.UnderlineString("Price"))
		columns = append(columns, table.ColumnConfig{
			Number:      i,
			Align:       text.AlignRight,
			AlignHeader: text.AlignRight,
		})
		i++
	}
	if contains(fields, "monthlyQuantity") {
		headers = append(headers, ui.UnderlineString("Monthly Qty"))
		columns = append(columns, table.ColumnConfig{
			Number:      i,
			Align:       text.AlignRight,
			AlignHeader: text.AlignRight,
		})
		i++
	}
	if contains(fields, "unit") {
		headers = append(headers, ui.UnderlineString("Unit"))
		columns = append(columns, table.ColumnConfig{
			Number:      i,
			Align:       text.AlignLeft,
			AlignHeader: text.AlignLeft,
		})
		i++
	}
	if contains(fields, "hourlyCost") {
		headers = append(headers, ui.UnderlineString("Hourly Cost"))
		columns = append(columns, table.ColumnConfig{
			Number:      i,
			Align:       text.AlignRight,
			AlignHeader: text.AlignRight,
		})
		i++
	}
	if contains(fields, "monthlyCost") {
		headers = append(headers, ui.UnderlineString("Monthly Cost"))
		columns = append(columns, table.ColumnConfig{
			Number:      i,
			Align:       text.AlignRight,
			AlignHeader: text.AlignRight,
		})
		i++
	}

	t.AppendRow(table.Row{""})

	t.SetColumnConfigs(columns)
	t.AppendHeader(headers)

	for _, r := range breakdown.Resources {
		t.AppendRow(table.Row{ui.BoldString(r.Name)})

		buildCostComponentRows(t, r.CostComponents, "", len(r.SubResources) > 0, fields)
		buildSubResourceRows(t, r.SubResources, "", fields)

		t.AppendRow(table.Row{""})
	}

	if includeTotal {
		var totalCostRow table.Row
		totalCostRow = append(totalCostRow, ui.BoldString("Project total"))
		numOfFields := i - 3
		for q := 0; q < numOfFields; q++ {
			totalCostRow = append(totalCostRow, "")
		}
		totalCostRow = append(totalCostRow, formatCost2DP(breakdown.TotalMonthlyCost))
		t.AppendRow(totalCostRow)
	}

	return t.Render()
}

func buildSubResourceRows(t table.Writer, subresources []Resource, prefix string, fields []string) {
	for i, r := range subresources {
		labelPrefix := prefix + "├─"
		nextPrefix := prefix + "│  "
		if i == len(subresources)-1 {
			labelPrefix = prefix + "└─"
			nextPrefix = prefix + "   "
		}

		t.AppendRow(table.Row{fmt.Sprintf("%s %s", ui.FaintString(labelPrefix), r.Name)})

		buildCostComponentRows(t, r.CostComponents, nextPrefix, len(r.SubResources) > 0, fields)
		buildSubResourceRows(t, r.SubResources, nextPrefix, fields)
	}
}

func buildCostComponentRows(t table.Writer, costComponents []CostComponent, prefix string, hasSubResources bool, fields []string) {
	for i, c := range costComponents {
		labelPrefix := prefix + "├─"
		if !hasSubResources && i == len(costComponents)-1 {
			labelPrefix = prefix + "└─"
		}

		label := fmt.Sprintf("%s %s", ui.FaintString(labelPrefix), c.Name)

		if c.MonthlyCost == nil {
			price := fmt.Sprintf("Monthly cost depends on usage: %s per %s",
				formatPrice(c.Price),
				c.Unit,
			)

			t.AppendRow(table.Row{
				label,
				price,
				price,
				price,
			}, table.RowConfig{AutoMerge: true, AlignAutoMerge: text.AlignLeft})
		} else {
			var tableRow table.Row
			tableRow = append(tableRow, label)

			if contains(fields, "price") {
				tableRow = append(tableRow, formatPrice(c.Price))
			}
			if contains(fields, "monthlyQuantity") {
				tableRow = append(tableRow, formatQuantity(c.MonthlyQuantity))
			}
			if contains(fields, "unit") {
				tableRow = append(tableRow, c.Unit)
			}
			if contains(fields, "hourlyCost") {
				tableRow = append(tableRow, formatCost2DP(c.HourlyCost))
			}
			if contains(fields, "monthlyCost") {
				tableRow = append(tableRow, formatCost2DP(c.MonthlyCost))
			}

			t.AppendRow(tableRow)
		}
	}
}
