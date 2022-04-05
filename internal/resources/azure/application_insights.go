package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"

	"github.com/shopspring/decimal"
)

type ApplicationInsights struct {
	Address               string
	Region                string
	RetentionInDays       int64
	MonthlyDataIngestedGB *float64 `infracost_usage:"monthly_data_ingested_gb"`
}

var ApplicationInsightsUsageSchema = []*schema.UsageItem{{Key: "monthly_data_ingested_gb", ValueType: schema.Float64, DefaultValue: 0}}

func (r *ApplicationInsights) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ApplicationInsights) BuildResource() *schema.Resource {
	region := r.Region
	costComponents := []*schema.CostComponent{}

	var dataIngested *decimal.Decimal
	if r.MonthlyDataIngestedGB != nil {
		dataIngested = decimalPtr(decimal.NewFromFloat(*r.MonthlyDataIngestedGB))
	}
	costComponents = append(costComponents, appInsightCostComponents(region, "Data ingested", "GB", "Enterprise Overage Data", "Enterprise", dataIngested))

	var dataRetentionDays *decimal.Decimal
	if r.RetentionInDays != 0 {
		dataRetentionDays = decimalPtr(decimal.NewFromInt(r.RetentionInDays))

		if dataRetentionDays.GreaterThan(decimal.NewFromInt(90)) && dataIngested != nil {
			days := dataRetentionDays.Sub(decimal.NewFromInt(90)).Div(decimal.NewFromInt(30))
			qty := decimalPtr(dataIngested.Mul(days))

			costComponents = append(costComponents, appInsightCostComponents(
				region,
				fmt.Sprintf("Data retention (%s days)", dataRetentionDays.String()),
				"GB",
				"Data Retention",
				"Enterprise",
				qty,
			))
		}
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents, UsageSchema: ApplicationInsightsUsageSchema,
	}
}

func appInsightCostComponents(region, name, unit, meterName, skuName string, qty *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            unit,
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Application Insights"),
			ProductFamily: strPtr("Management and Governance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", meterName))},
				{Key: "skuName", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", skuName))},
			},
		},
	}
}
