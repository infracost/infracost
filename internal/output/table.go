package output

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/urfave/cli/v2"

	"github.com/olekukonko/tablewriter"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

func ToTable(resources []*schema.Resource, c *cli.Context) ([]byte, error) {
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

	for _, r := range resources {
		if r.IsSkipped {
			continue
		}
		t.Append([]string{r.Name, "", "", "", "", ""})

		buildCostComponentRows(t, r.CostComponents, "", len(r.SubResources) > 0)
		buildSubResourceRows(t, r.SubResources, "")

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

	msg := skippedResourcesMessage(resources, c.Bool("show-skipped"))
	if msg != "" {
		_, err := bufw.WriteString(fmt.Sprintf("\n%s", msg))
		if err != nil {
			// The error here would just mean the output is shortened, so no need to return an error to the user in this case
			log.Errorf("Error writing skipped resources message: %v", err.Error())
		}
	}

	bufw.Flush()
	return buf.Bytes(), nil
}

func buildSubResourceRows(t *tablewriter.Table, subresources []*schema.Resource, prefix string) {
	color := []tablewriter.Colors{
		{tablewriter.FgHiBlackColor},
	}
	if config.Config.NoColor {
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

		buildCostComponentRows(t, r.CostComponents, nextPrefix, len(r.SubResources) > 0)
		buildSubResourceRows(t, r.SubResources, nextPrefix)
	}
}

func buildCostComponentRows(t *tablewriter.Table, costComponents []*schema.CostComponent, prefix string, hasSubResources bool) {
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

	for i, c := range costComponents {
		labelPrefix := prefix + "├─"
		if !hasSubResources && i == len(costComponents)-1 {
			labelPrefix = prefix + "└─"
		}

		t.Rich([]string{
			fmt.Sprintf("%s %s", labelPrefix, c.Name),
			formatQuantity(*c.MonthlyQuantity),
			c.Unit,
			formatCost(c.Price()),
			formatCost(c.HourlyCost()),
			formatCost(c.MonthlyCost()),
		}, color)
	}
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
