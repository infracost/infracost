package output

import (
	"bufio"
	"bytes"
	"fmt"

	"github.com/dustin/go-humanize"
	"github.com/olekukonko/tablewriter"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

func ToTableDeprecated(out Root, opts Options) ([]byte, error) {
	var buf bytes.Buffer
	bufw := bufio.NewWriter(&buf)

	t := tablewriter.NewWriter(bufw)
	t.SetHeader([]string{"NAME", "MONTHLY QTY", "UNIT", "PRICE", "HOURLY COST", "MONTHLY COST"})
	t.SetBorder(false)
	t.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	t.SetAutoWrapText(false)
	t.SetCenterSeparator("")
	t.SetColumnSeparator("")
	t.SetRowSeparator("")
	t.SetColumnAlignment([]int{
		tablewriter.ALIGN_LEFT,  // name
		tablewriter.ALIGN_RIGHT, // monthly quantity
		tablewriter.ALIGN_LEFT,  // unit
		tablewriter.ALIGN_RIGHT, // price
		tablewriter.ALIGN_RIGHT, // hourly cost
		tablewriter.ALIGN_RIGHT, // monthly cost
	})

	for _, r := range out.Resources {
		t.Append([]string{r.Name, "", "", "", "", ""})

		buildCostComponentRowsDeprecated(t, r.CostComponents, "", len(r.SubResources) > 0, opts.NoColor)
		buildSubResourceRowsDeprecated(t, r.SubResources, "", opts.NoColor)

		t.Append([]string{
			"Total",
			"",
			"",
			"",
			formatCostDeprecated(r.HourlyCost),
			formatCostDeprecated(r.MonthlyCost),
		})
		t.Append([]string{"", "", "", "", "", ""})
	}

	t.Append([]string{
		"OVERALL TOTAL (USD)",
		"",
		"",
		"",
		formatCostDeprecated(out.TotalHourlyCost),
		formatCostDeprecated(out.TotalMonthlyCost),
	})

	t.Render()

	msg := out.unsupportedResourcesMessage(opts.ShowSkipped)
	if msg != "" {
		_, err := bufw.WriteString(fmt.Sprintf("\n%s", msg))
		if err != nil {
			// The error here would just mean the output is shortened, so no need to return an error to the user in this case
			log.Errorf("Error writing unsupported resources message: %v", err.Error())
		}
	}

	bufw.Flush()
	return buf.Bytes(), nil
}

func buildSubResourceRowsDeprecated(t *tablewriter.Table, subresources []Resource, prefix string, noColor bool) {
	color := []tablewriter.Colors{
		{tablewriter.FgHiBlackColor},
	}
	if noColor {
		color = nil
	}

	for i, r := range subresources {
		labelPrefix := prefix + "├─"
		nextPrefix := prefix + "│  "
		if i == len(subresources)-1 {
			labelPrefix = prefix + "└─"
			nextPrefix = prefix + "   "
		}

		t.Rich([]string{fmt.Sprintf("%s %s", labelPrefix, r.Name)}, color)

		buildCostComponentRowsDeprecated(t, r.CostComponents, nextPrefix, len(r.SubResources) > 0, noColor)
		buildSubResourceRowsDeprecated(t, r.SubResources, nextPrefix, noColor)
	}
}

func buildCostComponentRowsDeprecated(t *tablewriter.Table, costComponents []CostComponent, prefix string, hasSubResources bool, noColor bool) {
	color := []tablewriter.Colors{
		{tablewriter.FgHiBlackColor},
		{tablewriter.FgHiBlackColor},
		{tablewriter.FgHiBlackColor},
		{tablewriter.FgHiBlackColor},
		{tablewriter.FgHiBlackColor},
		{tablewriter.FgHiBlackColor},
	}
	if noColor {
		color = nil
	}

	for i, c := range costComponents {
		labelPrefix := prefix + "├─"
		if !hasSubResources && i == len(costComponents)-1 {
			labelPrefix = prefix + "└─"
		}

		t.Rich([]string{
			fmt.Sprintf("%s %s", labelPrefix, c.Name),
			formatQuantity(c.MonthlyQuantity),
			c.Unit,
			formatCostDeprecated(&c.Price),
			formatCostDeprecated(c.HourlyCost),
			formatCostDeprecated(c.MonthlyCost),
		}, color)
	}
}

func formatCostDeprecated(d *decimal.Decimal) string {
	if d == nil {
		return "-"
	}

	f, _ := d.Float64()
	if f < 0.00005 && f != 0 {
		return fmt.Sprintf("%.g", f)
	}

	return humanize.FormatFloat("#,###.####", f)
}
