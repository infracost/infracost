package output

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"infracost/pkg/base"
	"infracost/pkg/config"

	"github.com/olekukonko/tablewriter"
	"github.com/shopspring/decimal"
)

func getLineItemCount(breakdown base.ResourceCostBreakdown) int {
	count := len(breakdown.PriceComponentCosts)

	for _, subResourceBreakdown := range breakdown.SubResourceCosts {
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

func ToTable(resourceCostBreakdowns []base.ResourceCostBreakdown) ([]byte, error) {
	var buf bytes.Buffer
	bufw := bufio.NewWriter(&buf)

	table := tablewriter.NewWriter(bufw)
	table.SetHeader([]string{"NAME", "HOURLY COST", "MONTHLY COST"})
	table.SetBorder(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT})

	overallTotalHourly := decimal.Zero
	overallTotalMonthly := decimal.Zero

	color := []tablewriter.Colors{
		tablewriter.Colors{tablewriter.FgHiBlackColor},
		tablewriter.Colors{tablewriter.FgHiBlackColor},
		tablewriter.Colors{tablewriter.FgHiBlackColor},
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

			row := []string{
				fmt.Sprintf("%s %s", getTreePrefix(lineItem, lineItemCount), priceComponentCost.PriceComponent.Name()),
				formatDecimal(priceComponentCost.HourlyCost, "%.4f"),
				formatDecimal(priceComponentCost.MonthlyCost, "%.4f"),
			}
			table.Rich(row, color)
		}

		for _, subResourceBreakdown := range breakdown.SubResourceCosts {
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
					formatDecimal(priceComponentCost.HourlyCost, "%.4f"),
					formatDecimal(priceComponentCost.MonthlyCost, "%.4f"),
				}
				table.Rich(row, color)
			}
		}

		table.Append([]string{
			"Total",
			formatDecimal(totalHourly, "%.4f"),
			formatDecimal(totalMonthly, "%.4f"),
		})
		table.Append([]string{"", "", ""})

		overallTotalHourly = overallTotalHourly.Add(totalHourly)
		overallTotalMonthly = overallTotalMonthly.Add(totalMonthly)
	}

	table.Append([]string{
		"OVERALL TOTAL",
		formatDecimal(overallTotalHourly, "%.4f"),
		formatDecimal(overallTotalMonthly, "%.4f"),
	})

	table.Render()

	bufw.Flush()
	return buf.Bytes(), nil
}
