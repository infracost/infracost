package output

import (
	"fmt"

	"github.com/infracost/infracost/internal/ui"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func ToTable(out Root, opts Options) ([]byte, error) {
	s := ""

	hasNilCosts := false

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

		s += tableForBreakdown(*project.Breakdown, opts.Fields)
		s += "\n"

		if i != len(out.Projects)-1 {
			s += "\n"
		}
	}

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

func tableForBreakdown(breakdown Breakdown, fields []string) string {
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
	if contains(fields, "monthly_quantity") {
		headers = append(headers, ui.UnderlineString("Quantity"))
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
	if contains(fields, "hourly_cost") {
		headers = append(headers, ui.UnderlineString("Hourly Cost"))
		columns = append(columns, table.ColumnConfig{
			Number:      i,
			Align:       text.AlignRight,
			AlignHeader: text.AlignRight,
		})
		i++
	}
	if contains(fields, "monthly_cost") {
		headers = append(headers, ui.UnderlineString("Monthly Cost"))
		columns = append(columns, table.ColumnConfig{
			Number:      i,
			Align:       text.AlignRight,
			AlignHeader: text.AlignRight,
		})
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

	var totalCostRow table.Row
	totalCostRow = append(totalCostRow, ui.BoldString("PROJECT TOTAL"))
	numOfFields := len(fields)
	if contains(fields, "name") {
		numOfFields--
	}
	for q := 1; q < numOfFields; q++ {
		totalCostRow = append(totalCostRow, "")
	}
	totalCostRow = append(totalCostRow, formatCost2DP(breakdown.TotalMonthlyCost))
	t.AppendRow(totalCostRow)

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
			price := fmt.Sprintf("Cost depends on usage: %s per %s",
				formatPrice(c.Price),
				c.Unit,
			)

			t.AppendRow(table.Row{
				label,
				ui.FaintString(price),
				ui.FaintString(price),
				ui.FaintString(price),
			}, table.RowConfig{AutoMerge: true, AlignAutoMerge: text.AlignLeft})
		} else {
			var tableRow table.Row
			tableRow = append(tableRow, label)

			if contains(fields, "price") {
				tableRow = append(tableRow, formatPrice(c.Price))
			}
			if contains(fields, "monthly_quantity") {
				tableRow = append(tableRow, formatQuantity(c.MonthlyQuantity))
			}
			if contains(fields, "unit") {
				tableRow = append(tableRow, c.Unit)
			}
			if contains(fields, "hourly_cost") {
				tableRow = append(tableRow, formatCost2DP(c.HourlyCost))
			}
			if contains(fields, "monthly_cost") {
				tableRow = append(tableRow, formatCost2DP(c.MonthlyCost))
			}

			t.AppendRow(tableRow)
		}
	}
}
