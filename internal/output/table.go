package output

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"

	"github.com/infracost/infracost/internal/ui"

	log "github.com/sirupsen/logrus"
)

func ToTable(out Root, opts Options) ([]byte, error) {
	var tableLen int

	s := ""

	// Don't show the project total if there's only one project result
	// since we will show the overall total anyway
	includeProjectTotals := len(out.Projects) != 1

	for i, project := range out.Projects {
		if project.Breakdown == nil {
			continue
		}

		if i != 0 {
			s += "──────────────────────────────────\n"
		}

		s += fmt.Sprintf("%s %s\n",
			ui.BoldString("Project:"),
			project.Label(),
		)

		if project.Metadata.TerraformModulePath != "" {
			s += fmt.Sprintf("%s %s\n",
				ui.BoldString("Module path:"),
				project.Metadata.TerraformModulePath,
			)
		}

		if project.Metadata.WorkspaceLabel() != "" {
			s += fmt.Sprintf("%s %s\n",
				ui.BoldString("Workspace:"),
				project.Metadata.WorkspaceLabel(),
			)
		}

		s += "\n"

		tableOut := tableForBreakdown(out.Currency, *project.Breakdown, opts.Fields, includeProjectTotals)

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

	totalOut := formatCost2DP(out.Currency, out.TotalMonthlyCost)

	overallTitle := formatTitleWithCurrency(" OVERALL TOTAL", out.Currency)
	s += fmt.Sprintf("%s%s",
		ui.BoldString(overallTitle),
		fmt.Sprintf("%*s ", tableLen-(len(overallTitle)+1), totalOut), // pad based on the last line length
	)

	summaryMsg := out.summaryMessage(opts.ShowSkipped)

	if summaryMsg != "" {
		s += "\n──────────────────────────────────\n" + summaryMsg
	}

	return []byte(s), nil
}

func tableForBreakdown(currency string, breakdown Breakdown, fields []string, includeTotal bool) string {
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
		headers = append(headers, ui.UnderlineString(formatTitleWithCurrency("Price", currency)))
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
		headers = append(headers, ui.UnderlineString(formatTitleWithCurrency("Hourly Cost", currency)))
		columns = append(columns, table.ColumnConfig{
			Number:      i,
			Align:       text.AlignRight,
			AlignHeader: text.AlignRight,
		})
		i++
	}
	if contains(fields, "monthlyCost") {
		headers = append(headers, ui.UnderlineString(formatTitleWithCurrency("Monthly Cost", currency)))
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
		filteredComponents := filterZeroValComponents(r.CostComponents, r.Name)
		filteredSubResources := filterZeroValResources(r.SubResources, r.Name)
		if len(filteredComponents) == 0 && len(filteredSubResources) == 0 {
			log.Info(fmt.Sprintf("Hiding resource with no usage: %s", r.Name))
			continue
		}

		t.AppendRow(table.Row{ui.BoldString(r.Name)})

		buildCostComponentRows(t, currency, filteredComponents, "", len(r.SubResources) > 0, fields)
		buildSubResourceRows(t, currency, filteredSubResources, "", fields)

		t.AppendRow(table.Row{""})
	}

	if includeTotal {
		var totalCostRow table.Row
		totalCostRow = append(totalCostRow, ui.BoldString(formatTitleWithCurrency("Project total", currency)))
		numOfFields := i - 3
		for q := 0; q < numOfFields; q++ {
			totalCostRow = append(totalCostRow, "")
		}
		totalCostRow = append(totalCostRow, formatCost2DP(currency, breakdown.TotalMonthlyCost))
		t.AppendRow(totalCostRow)
	}

	return t.Render()
}

func buildSubResourceRows(t table.Writer, currency string, subresources []Resource, prefix string, fields []string) {
	for i, r := range subresources {
		filteredComponents := filterZeroValComponents(r.CostComponents, r.Name)
		filteredSubResources := filterZeroValResources(r.SubResources, r.Name)
		if len(filteredComponents) == 0 && len(filteredSubResources) == 0 {
			continue
		}

		labelPrefix := prefix + "├─"
		nextPrefix := prefix + "│  "
		if i == len(subresources)-1 {
			labelPrefix = prefix + "└─"
			nextPrefix = prefix + "   "
		}

		t.AppendRow(table.Row{fmt.Sprintf("%s %s", ui.FaintString(labelPrefix), r.Name)})

		buildCostComponentRows(t, currency, filteredComponents, nextPrefix, len(r.SubResources) > 0, fields)
		buildSubResourceRows(t, currency, filteredSubResources, nextPrefix, fields)
	}
}

func buildCostComponentRows(t table.Writer, currency string, costComponents []CostComponent, prefix string, hasSubResources bool, fields []string) {
	for i, c := range costComponents {
		labelPrefix := prefix + "├─"
		if !hasSubResources && i == len(costComponents)-1 {
			labelPrefix = prefix + "└─"
		}

		label := fmt.Sprintf("%s %s", ui.FaintString(labelPrefix), c.Name)

		if c.MonthlyCost == nil {
			price := fmt.Sprintf("Monthly cost depends on usage: %s per %s",
				formatPrice(currency, c.Price),
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
				tableRow = append(tableRow, formatPrice(currency, c.Price))
			}
			if contains(fields, "monthlyQuantity") {
				tableRow = append(tableRow, formatQuantity(c.MonthlyQuantity))
			}
			if contains(fields, "unit") {
				tableRow = append(tableRow, c.Unit)
			}
			if contains(fields, "hourlyCost") {
				tableRow = append(tableRow, formatCost2DP(currency, c.HourlyCost))
			}
			if contains(fields, "monthlyCost") {
				tableRow = append(tableRow, formatCost2DP(currency, c.MonthlyCost))
			}

			t.AppendRow(tableRow)
		}
	}
}

func filterZeroValComponents(costComponents []CostComponent, resourceName string) []CostComponent {
	var filteredComponents []CostComponent
	for _, c := range costComponents {
		if c.MonthlyQuantity != nil && c.MonthlyQuantity.IsZero() {
			log.Info(fmt.Sprintf("Hiding cost with no usage: %s '%s'", resourceName, c.Name))
			continue
		}

		filteredComponents = append(filteredComponents, c)
	}
	return filteredComponents
}

func filterZeroValResources(resources []Resource, resourceName string) []Resource {
	var filteredResources []Resource
	for _, r := range resources {
		filteredComponents := filterZeroValComponents(r.CostComponents, fmt.Sprintf("%s.%s", resourceName, r.Name))
		filteredSubResources := filterZeroValResources(r.SubResources, fmt.Sprintf("%s.%s", resourceName, r.Name))
		if len(filteredComponents) == 0 && len(filteredSubResources) == 0 {
			log.Info(fmt.Sprintf("Hiding resource with no usage: %s.%s", resourceName, r.Name))
			continue
		}

		filteredResources = append(filteredResources, r)
	}
	return filteredResources
}
