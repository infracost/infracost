package output

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"

	"github.com/infracost/infracost/pkg/config"
	"github.com/infracost/infracost/pkg/schema"

	"github.com/olekukonko/tablewriter"
	"github.com/shopspring/decimal"
)

func getLineItemCount(r *schema.Resource) int {
	c := len(r.CostComponents)

	for _, s := range r.FlattenedSubResources() {
		c += len(s.CostComponents)
	}

	return c
}

func getTreePrefix(lineItem int, lineItemCount int) string {
	if lineItem == lineItemCount {
		return "└─"
	}

	return "├─"
}

func formatCost(d decimal.Decimal) string {
	f, _ := d.Float64()
	if f < 0.00005 && f != 0 {
		return fmt.Sprintf("%.g", f)
	}

	return fmt.Sprintf("%.4f", f)
}

func formatQuantity(quantity decimal.Decimal) string {
	f, _ := quantity.Float64()
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func ToTable(resources []*schema.Resource) ([]byte, error) {
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

	overallTotalHourly := decimal.Zero
	overallTotalMonthly := decimal.Zero

	color := []tablewriter.Colors{
		{tablewriter.FgHiBlackColor},
		{tablewriter.FgHiBlackColor},
		{tablewriter.FgHiBlackColor},
		{tablewriter.FgHiBlackColor},
		{tablewriter.FgHiBlackColor},
		{tablewriter.FgHiBlackColor},
	}
	if config.Config.NoColor {
		color = nil
	}

	for _, r := range resources {
		if r.IsSkipped {
			continue
		}
		t.Append([]string{r.Name, "", "", "", "", ""})

		lineItemCount := getLineItemCount(r)
		lineItem := 0

		for _, c := range r.CostComponents {
			lineItem++

			row := []string{
				fmt.Sprintf("%s %s", getTreePrefix(lineItem, lineItemCount), c.Name),
				formatQuantity(*c.MonthlyQuantity),
				c.Unit,
				formatCost(c.Price()),
				formatCost(c.HourlyCost()),
				formatCost(c.MonthlyCost()),
			}
			t.Rich(row, color)
		}

		for _, s := range r.FlattenedSubResources() {
			for _, c := range s.CostComponents {
				lineItem++

				row := []string{
					fmt.Sprintf("%s %s (%s)", getTreePrefix(lineItem, lineItemCount), c.Name, s.Name),
					formatQuantity(*c.MonthlyQuantity),
					c.Unit,
					formatCost(c.Price()),
					formatCost(c.HourlyCost()),
					formatCost(c.MonthlyCost()),
				}
				t.Rich(row, color)
			}
		}

		t.Append([]string{
			"Total",
			"",
			"",
			"",
			formatCost(r.HourlyCost()),
			formatCost(r.MonthlyCost()),
		})
		t.Append([]string{"", "", "", "", "", ""})

		overallTotalHourly = overallTotalHourly.Add(r.HourlyCost())
		overallTotalMonthly = overallTotalMonthly.Add(r.MonthlyCost())
	}

	t.Append([]string{
		"OVERALL TOTAL",
		"",
		"",
		"",
		formatCost(overallTotalHourly),
		formatCost(overallTotalMonthly),
	})

	t.Render()

	bufw.Flush()
	return buf.Bytes(), nil
}
