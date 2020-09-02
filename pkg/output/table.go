package output

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"

	"infracost/pkg/config"
	"infracost/pkg/schema"

	"github.com/olekukonko/tablewriter"
	"github.com/shopspring/decimal"
)

func getLineItemCount(resource *schema.Resource) int {
	count := len(resource.CostComponents)

	for _, subResource := range resource.FlattenedSubResources() {
		count += len(subResource.CostComponents)
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

func ToTable(resources []*schema.Resource) ([]byte, error) {
	var buf bytes.Buffer
	bufw := bufio.NewWriter(&buf)

	table := tablewriter.NewWriter(bufw)
	table.SetHeader([]string{"NAME", "MONTHLY QTY", "UNIT", "PRICE", "HOURLY COST", "MONTHLY COST"})
	table.SetBorder(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetColumnAlignment([]int{
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

	for _, resource := range resources {
		table.Append([]string{resource.Name, "", ""})

		lineItemCount := getLineItemCount(resource)
		lineItem := 0

		for _, costComponent := range resource.CostComponents {
			lineItem++

			row := []string{
				fmt.Sprintf("%s %s", getTreePrefix(lineItem, lineItemCount), costComponent.Name),
				formatQuantity(*costComponent.MonthlyQuantity),
				costComponent.Unit,
				formatDecimal(costComponent.Price(), "%.4f"),
				formatDecimal(costComponent.HourlyCost(), "%.4f"),
				formatDecimal(costComponent.MonthlyCost(), "%.4f"),
			}
			table.Rich(row, color)
		}

		for _, subResource := range resource.FlattenedSubResources() {
			for _, costComponent := range subResource.CostComponents {
				lineItem++

				row := []string{
					fmt.Sprintf("%s %s (%s)", getTreePrefix(lineItem, lineItemCount), costComponent.Name, subResource.Name),
					formatQuantity(*costComponent.MonthlyQuantity),
					costComponent.Unit,
					formatDecimal(costComponent.Price(), "%.4f"),
					formatDecimal(costComponent.HourlyCost(), "%.4f"),
					formatDecimal(costComponent.MonthlyCost(), "%.4f"),
				}
				table.Rich(row, color)
			}
		}

		table.Append([]string{
			"Total",
			"",
			"",
			"",
			formatDecimal(resource.HourlyCost(), "%.4f"),
			formatDecimal(resource.MonthlyCost(), "%.4f"),
		})
		table.Append([]string{"", "", ""})

		overallTotalHourly = overallTotalHourly.Add(resource.HourlyCost())
		overallTotalMonthly = overallTotalMonthly.Add(resource.MonthlyCost())
	}

	table.Append([]string{
		"OVERALL TOTAL",
		"",
		"",
		"",
		formatDecimal(overallTotalHourly, "%.4f"),
		formatDecimal(overallTotalMonthly, "%.4f"),
	})

	table.Render()

	bufw.Flush()
	return buf.Bytes(), nil
}
