package output

import (
	"fmt"

	"github.com/infracost/infracost/internal/ui"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type TableColumns struct {
	Name            bool
	Price           bool
	MonthlyQuantity bool
	Unit            bool
	HourlyCost      bool
	MonthlyCost     bool
}

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

	fmt.Println(opts.Fields)

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
	columnName := map[string]string{
		"name":             "Name",
		"monthly_quantity": "Quantity",
		"price":            "Price",
		"unit":             "Unit",
		"hourly_cost":      "Hourly Cost",
		"monthly_cost":     "Monthly Cost",
	}

	fmt.Println(columnName)

	t := table.NewWriter()
	t.Style().Options.DrawBorder = false
	t.Style().Options.SeparateColumns = false
	t.Style().Options.SeparateRows = false
	t.Style().Options.SeparateHeader = false
	t.Style().Format.Header = text.FormatDefault

	var columns []table.ColumnConfig
	var headers table.Row
	columnNumber := len(fields)

	for i := 1; i < columnNumber+1; i++ {
		if i%2 != 0 {
			columns = append(columns, table.ColumnConfig{
				Number:      i,
				Align:       text.AlignLeft,
				AlignHeader: text.AlignLeft,
			})
			headers = append(headers, table.Row{
				ui.UnderlineString(fields[i-1]),
			})
		} else {
			columns = append(columns, table.ColumnConfig{
				Number:      i,
				Align:       text.AlignRight,
				AlignHeader: text.AlignRight,
			})
			headers = append(headers, table.Row{
				ui.UnderlineString(fields[i-1]),
			})
		}
	}

	t.SetColumnConfigs(columns)
	t.AppendHeader(headers)

	t.AppendRow(table.Row{"", "", "", ""})

	for _, r := range breakdown.Resources {
		t.AppendRow(table.Row{ui.BoldString(r.Name), "", "", ""})

		buildCostComponentRows(t, r.CostComponents, "", len(r.SubResources) > 0)
		buildSubResourceRows(t, r.SubResources, "")

		t.AppendRow(table.Row{"", "", "", ""})
	}

	t.AppendRow(table.Row{
		ui.BoldString("PROJECT TOTAL"),
		"",
		"",
		formatCost2DP(breakdown.TotalMonthlyCost),
	})

	return t.Render()
}

func buildSubResourceRows(t table.Writer, subresources []Resource, prefix string) {
	for i, r := range subresources {
		labelPrefix := prefix + "├─"
		nextPrefix := prefix + "│  "
		if i == len(subresources)-1 {
			labelPrefix = prefix + "└─"
			nextPrefix = prefix + "   "
		}

		t.AppendRow(table.Row{fmt.Sprintf("%s %s", ui.FaintString(labelPrefix), r.Name)})

		buildCostComponentRows(t, r.CostComponents, nextPrefix, len(r.SubResources) > 0)
		buildSubResourceRows(t, r.SubResources, nextPrefix)
	}
}

func buildCostComponentRows(t table.Writer, costComponents []CostComponent, prefix string, hasSubResources bool) {
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
			t.AppendRow(table.Row{
				label,
				formatQuantity(c.MonthlyQuantity),
				c.Unit,
				formatCost2DP(c.MonthlyCost),
			})
		}
	}
}
