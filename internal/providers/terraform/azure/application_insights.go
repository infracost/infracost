package azure

import (
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMApplicationInsightsRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_application_insights",
		RFunc: NewAzureRMApplicationInsights,
	}
}

func NewAzureRMApplicationInsights(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{})
	costComponents := []*schema.CostComponent{}

	var dataIngested *decimal.Decimal
	if u != nil && u.Get("monthly_data_ingested_gb").Type != gjson.Null {
		dataIngested = decimalPtr(decimal.NewFromInt(u.Get("monthly_data_ingested_gb").Int()))
	}
	costComponents = append(costComponents, appInsightCostComponents(region, "Data ingested", "GB", "Enterprise Overage Data", "Enterprise", dataIngested))

	var dataRetentionDays *decimal.Decimal
	if d.Get("retention_in_days").Type != gjson.Null {
		dataRetentionDays = decimalPtr(decimal.NewFromInt(d.Get("retention_in_days").Int()))

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
		Name:           d.Address,
		CostComponents: costComponents,
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
