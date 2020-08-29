package output

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"infracost/pkg/config"
	"infracost/pkg/costs"

	"github.com/olekukonko/tablewriter"
	"github.com/shopspring/decimal"
)

func getLineItemCount(breakdown costs.ResourceCostBreakdown) int {
	count := len(breakdown.PriceComponentCosts)

	for _, subResourceBreakdown := range flattenSubResourceBreakdowns(breakdown.SubResourceCosts) {
		count += len(subResourceBreakdown.PriceComponentCosts)
	}

	return count
}

func getTreePrefix(lineItem int, lineItemCount int) string {
	if lineItem == lineItemCount {
		return "└─"
	}
	return "├─"
}

func formatDecimal(d decimal.Decimal, format string) string {
	f, _ := d.Float64()
	return fmt.Sprintf(format, f)
}

func formatQuantity(quantity decimal.Decimal) string {
	f, _ := quantity.Float64()
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func flattenSubResourceBreakdowns(breakdowns []costs.ResourceCostBreakdown) []costs.ResourceCostBreakdown {
	flattenedBreakdowns := make([]costs.ResourceCostBreakdown, 0)
	for _, breakdown := range breakdowns {
		flattenedBreakdowns = append(flattenedBreakdowns, breakdown)
		if len(breakdown.SubResourceCosts) > 0 {
			flattenedBreakdowns = append(flattenedBreakdowns, breakdown.SubResourceCosts...)
		}
	}
	return flattenedBreakdowns
}

func ToTable(resourceCostBreakdowns []costs.ResourceCostBreakdown) ([]byte, error) {
	var buf bytes.Buffer
	bufw := bufio.NewWriter(&buf)

	table := tablewriter.NewWriter(bufw)
	table.SetHeader([]string{"NAME", "QUANTITY", "UNIT", "HOURLY COST", "MONTHLY COST"})
	table.SetBorder(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT})

	overallTotalHourly := decimal.Zero
	overallTotalMonthly := decimal.Zero

	color := []tablewriter.Colors{
		{tablewriter.FgHiBlackColor},
		{tablewriter.FgHiBlackColor},
		{tablewriter.FgHiBlackColor},
		{tablewriter.FgHiBlackColor},
		{tablewriter.FgHiBlackColor},
	}
	if config.Config.NoColor {
		color = nil
	}

	for _, breakdown := range resourceCostBreakdowns {
		table.Append([]string{breakdown.Resource.Address(), "", ""})

		lineItemCount := getLineItemCount(breakdown)
		lineItem := 0
		totalHourly := decimal.Zero
		totalMonthly := decimal.Zero

		for _, priceComponentCost := range breakdown.PriceComponentCosts {
			lineItem++

			totalHourly = totalHourly.Add(priceComponentCost.HourlyCost)
			totalMonthly = totalMonthly.Add(priceComponentCost.MonthlyCost)

			hourlyCost := formatDecimal(priceComponentCost.HourlyCost, "%.4f")
			if len(priceComponentCost.CostOverridingLabel) > 0 {
				hourlyCost = priceComponentCost.CostOverridingLabel
			}

			monthlyCost := formatDecimal(priceComponentCost.MonthlyCost, "%.4f")
			if len(priceComponentCost.CostOverridingLabel) > 0 {
				monthlyCost = priceComponentCost.CostOverridingLabel
			}

			row := []string{
				fmt.Sprintf("%s %s", getTreePrefix(lineItem, lineItemCount), priceComponentCost.PriceComponent.Name()),
				formatQuantity(priceComponentCost.PriceComponent.Quantity()),
				priceComponentCost.PriceComponent.Unit(),
				hourlyCost,
				monthlyCost,
			}
			table.Rich(row, color)
		}

		for _, subResourceBreakdown := range flattenSubResourceBreakdowns(breakdown.SubResourceCosts) {
			for _, priceComponentCost := range subResourceBreakdown.PriceComponentCosts {
				lineItem++

				totalHourly = totalHourly.Add(priceComponentCost.HourlyCost)
				totalMonthly = totalMonthly.Add(priceComponentCost.MonthlyCost)

				prefixToRemove := fmt.Sprintf("%s.", breakdown.Resource.Address())
				label := fmt.Sprintf("%s %s",
					strings.TrimPrefix(subResourceBreakdown.Resource.Address(), prefixToRemove),
					priceComponentCost.PriceComponent.Name(),
				)
				row := []string{
					fmt.Sprintf("%s %s", getTreePrefix(lineItem, lineItemCount), label),
					formatQuantity(priceComponentCost.PriceComponent.Quantity()),
					priceComponentCost.PriceComponent.Unit(),
					formatDecimal(priceComponentCost.HourlyCost, "%.4f"),
					formatDecimal(priceComponentCost.MonthlyCost, "%.4f"),
				}
				table.Rich(row, color)
			}
		}

		table.Append([]string{
			"Total",
			"",
			"",
			formatDecimal(totalHourly, "%.4f"),
			formatDecimal(totalMonthly, "%.4f"),
		})
		table.Append([]string{"", "", ""})

		overallTotalHourly = overallTotalHourly.Add(totalHourly)
		overallTotalMonthly = overallTotalMonthly.Add(totalMonthly)
	}

	table.Append([]string{
		"OVERALL TOTAL",
		"",
		"",
		formatDecimal(overallTotalHourly, "%.4f"),
		formatDecimal(overallTotalMonthly, "%.4f"),
	})

	table.Render()

	bufw.Flush()
	return buf.Bytes(), nil
}
