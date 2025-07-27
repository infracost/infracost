package output

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/ui"
)

func ToTable(out Root, opts Options) ([]byte, error) {
	var tableLen int

	hasUsageFootnote := false
	for _, f := range opts.Fields {
		if f == "monthlyQuantity" || f == "hourlyCost" || f == "monthlyCost" {
			hasUsageFootnote = true
			break
		}
	}

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

		s += projectTitle(project)
		s += "\n"

		if project.Metadata.HasErrors() {
			s += erroredProject(project)

			if len(out.Projects) == 1 {
				s += "\n"
			}
		} else {
			fields := opts.Fields
			if hasUsageFootnote {
				fields = append(fields, "usageFootnote")
			}
			tableOut := tableForBreakdown(out.Currency, *project.Breakdown, fields, includeProjectTotals)

			// Get the last table length so we can align the overall total with it
			if i == len(out.Projects)-1 {
				tableLen = len(ui.StripColor(strings.SplitN(tableOut, "\n", 2)[0]))
				if hasUsageFootnote {
					tableLen -= 3
				}
			}

			s += tableOut

			s += "\n"
		}

		if i != len(out.Projects)-1 {
			s += "\n"
		}
	}

	if includeProjectTotals {
		s += "\n"
	}

	totalOut := FormatCost2DP(out.Currency, out.TotalMonthlyCost)

	overallTitle := formatTitleWithCurrency(" OVERALL TOTAL", out.Currency)
	padding := 12
	if tableLen > 0 {
		padding = tableLen - (len(overallTitle) + 1)
	}

	s += fmt.Sprintf("%s%s",
		ui.BoldString(overallTitle),
		fmt.Sprintf("%*s ", padding, totalOut), // pad based on the last line length
	)

	if hasUsageFootnote {
		s += "\n\n"
		s += usageCostsMessage(out, false)
		s += "\n"
	}

	// Show missing prices warning prominently in table output
	if out.MissingPricesCount > 0 {
		if !hasUsageFootnote {
			s += "\n"
		}
		warningMsg := ui.WarningString("WARNING:")
		if out.MissingPricesCount == 1 {
			s += fmt.Sprintf("\n%s 1 price missing, cost estimates may be incomplete.\n", warningMsg)
		} else {
			s += fmt.Sprintf("\n%s %d prices missing, cost estimates may be incomplete.\n", warningMsg, out.MissingPricesCount)
		}
	}

	summaryMsg := out.summaryMessage(opts.ShowSkipped)

	if summaryMsg != "" {
		s += "\n──────────────────────────────────\n" + summaryMsg
	}

	if len(out.Projects) > 0 {
		s += "\n\n"
		s += breakdownSummaryTable(out, opts)
	}

	return []byte(s), nil
}

func erroredProject(project Project) string {
	s := ui.BoldString("Errors:") + "\n"

	for _, diag := range project.Metadata.Errors {
		pieces := strings.Split(diag.Message, ": ")
		for x, piece := range pieces {
			s += strings.Repeat("  ", x+1) + piece

			if len(pieces)-1 == x {
				s += "\n"
			} else {
				s += ":\n"
			}
		}
	}

	return s
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

	if contains(fields, "usageFootnote") {
		headers = append(headers, "")
		columns = append(columns, table.ColumnConfig{
			Number:      i,
			Align:       text.AlignLeft,
			AlignHeader: text.AlignLeft,
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
			logging.Logger.Debug().Msgf("Hiding resource with no usage: %s", r.Name)
			continue
		}

		t.AppendRow(table.Row{ui.BoldString(r.Name)})

		buildCostComponentRows(t, currency, filteredComponents, "", len(r.SubResources) > 0, fields)
		buildSubResourceRows(t, currency, filteredSubResources, "", fields)
		buildActualCostRows(t, currency, r.ActualCosts, "", fields)

		t.AppendRow(table.Row{""})
	}

	if includeTotal {
		var totalCostRow table.Row
		totalCostRow = append(totalCostRow, ui.BoldString(formatTitleWithCurrency("Project total", currency)))
		numOfFields := i - 3
		if contains(fields, "usageFootnote") {
			numOfFields -= 1
		}
		for q := 0; q < numOfFields; q++ {
			totalCostRow = append(totalCostRow, "")
		}
		totalCostRow = append(totalCostRow, FormatCost2DP(currency, breakdown.TotalMonthlyCost))
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
		buildActualCostRows(t, currency, r.ActualCosts, nextPrefix, fields)
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

			if c.PriceNotFound {
				price = "not found"
			}

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
				if c.PriceNotFound {
					tableRow = append(tableRow, "not found")
				} else {
					tableRow = append(tableRow, formatPrice(currency, c.Price))
				}
			}
			if contains(fields, "monthlyQuantity") {
				tableRow = append(tableRow, formatQuantity(c.MonthlyQuantity))
			}
			if contains(fields, "unit") {
				tableRow = append(tableRow, c.Unit)
			}
			if contains(fields, "hourlyCost") {
				if c.PriceNotFound {
					tableRow = append(tableRow, "not found")
				} else {
					tableRow = append(tableRow, FormatCost2DP(currency, c.HourlyCost))
				}
			}
			if contains(fields, "monthlyCost") {
				if c.PriceNotFound {
					tableRow = append(tableRow, "not found")
				} else {
					tableRow = append(tableRow, FormatCost2DP(currency, c.MonthlyCost))
				}
			}

			if contains(fields, "usageFootnote") {
				if c.UsageBased {
					tableRow = append(tableRow, "*")
				}
			}

			t.AppendRow(tableRow)
		}
	}
}

func buildActualCostRows(t table.Writer, currency string, actualCosts []ActualCosts, prefix string, fields []string) {
	for i, ac := range actualCosts {
		labelPrefix := prefix + "├─"
		if i == len(actualCosts)-1 {
			labelPrefix = prefix + "└─"
		}

		var dateRange string
		if !ac.StartTimestamp.IsZero() && !ac.EndTimestamp.IsZero() {
			// We want to display the range as "days" which means "inclusive", so subtract
			// 1 nano second from the exclusive endTimestamp.  This means the (exclusive) timestamp
			// range "2020/10/10 00:00:00Z-2020/10/20 00:00:00Z" will be displayed as the (inclusive)
			// day range "2020/10/10 - 2020/10/19".
			endDay := ac.EndTimestamp.Add(-1)
			endFmt := "Jan 2"
			if ac.StartTimestamp.Month() == endDay.Month() {
				endFmt = "2"
			}
			dateRange = fmt.Sprintf(" %s-%s", ac.StartTimestamp.Format("Jan 2"), endDay.Format(endFmt))
		}
		var resourceID string
		if ac.ResourceID != "" {
			resourceID = fmt.Sprintf(" (%s)", ac.ResourceID)
		}

		t.AppendRow(
			table.Row{fmt.Sprintf("%s Actual costs%s%s", labelPrefix, dateRange, resourceID)},
			table.RowConfig{AutoMerge: true},
		)
		buildCostComponentRows(t, currency, ac.CostComponents, prefix+"   ", false, fields)
	}
}

func filterZeroValComponents(costComponents []CostComponent, resourceName string) []CostComponent {
	filteredComponents := make([]CostComponent, 0, len(costComponents))
	for _, c := range costComponents {
		if c.MonthlyQuantity != nil && c.MonthlyQuantity.IsZero() {
			logging.Logger.Debug().Msgf("Hiding cost with no usage: %s '%s'", resourceName, c.Name)
			continue
		}

		filteredComponents = append(filteredComponents, c)
	}
	return filteredComponents
}

func filterZeroValResources(resources []Resource, resourceName string) []Resource {
	filteredResources := make([]Resource, 0, len(resources))
	for _, r := range resources {
		filteredComponents := filterZeroValComponents(r.CostComponents, fmt.Sprintf("%s.%s", resourceName, r.Name))
		filteredSubResources := filterZeroValResources(r.SubResources, fmt.Sprintf("%s.%s", resourceName, r.Name))
		if len(filteredComponents) == 0 && len(filteredSubResources) == 0 {
			logging.Logger.Debug().Msgf("Hiding resource with no usage: %s.%s", resourceName, r.Name)
			continue
		}

		filteredResources = append(filteredResources, r)
	}
	return filteredResources
}

func breakdownSummaryTable(out Root, _ Options) string {
	t := table.NewWriter()
	t.SetStyle(table.StyleBold)
	t.Style().Format.Header = text.FormatDefault
	t.AppendHeader(table.Row{
		"Project",
		"Baseline cost",
		"Usage cost*",
		"Total cost",
	})

	t.SetColumnConfigs([]table.ColumnConfig{
		{Name: "Project", WidthMin: 50},
		{Name: "Baseline cost", WidthMin: 10, Align: text.AlignRight},
		{Name: "Usage cost*", WidthMin: 10, Align: text.AlignRight},
		{Name: "Total cost", WidthMin: 10, Align: text.AlignRight},
	})

	for _, project := range out.Projects {
		baseline := project.Breakdown.TotalMonthlyCost
		if baseline != nil && project.Breakdown.TotalMonthlyUsageCost != nil {
			baseline = decimalPtr(baseline.Sub(*project.Breakdown.TotalMonthlyUsageCost))
		}

		t.AppendRow(
			table.Row{
				truncateMiddle(project.Label(), 64, "..."),
				formatCost(out.Currency, baseline),
				formatUsageCost(out, project.Breakdown.TotalMonthlyUsageCost),
				formatCost(out.Currency, project.Breakdown.TotalMonthlyCost),
			},
		)
	}

	return t.Render()
}
