package google

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"
)

func GetMonitoringItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_monitoring_metric_descriptor",
		RFunc: NewMonitoring,
	}
}

func NewMonitoring(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	costComponents := []*schema.CostComponent{}

	var monitData *decimal.Decimal
	if u != nil && u.Get("monthly_monitoring_data_mib").Exists() {
		monitData = decimalPtr(decimal.NewFromInt(u.Get("monthly_monitoring_data_mib").Int()))

		apiCallsLimits := []int{100000, 150000}

		apiCallsTiers := usage.CalculateTierBuckets(*monitData, apiCallsLimits)

		if apiCallsTiers[0].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, monitoringCostComponent("Monitoring data (first 100K)", "MiB-months", "150", "Metric Volume", &apiCallsTiers[0]))
		}

		if apiCallsTiers[1].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, monitoringCostComponent("Monitoring data (next 150K)", "MiB-months", "100000", "Metric Volume", &apiCallsTiers[1]))
		}

		if apiCallsTiers[2].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, monitoringCostComponent("Monitoring data (over 250K)", "MiB-months", "250000", "Metric Volume", &apiCallsTiers[2]))
		}
	} else {
		var unknown *decimal.Decimal
		costComponents = append(costComponents, monitoringCostComponent("Monitoring data (first 100K)", "MiB-months", "150", "Metric Volume", unknown))
	}

	var monitAPI *decimal.Decimal
	if u != nil && u.Get("monthly_monitoring_api_calls").Exists() {
		monitAPI = decimalPtr(decimal.NewFromInt(u.Get("monthly_monitoring_api_calls").Int()))
	}

	costComponents = append(costComponents, monitoringCostComponent("Monitoring API calls", "calls", "1000000", "Monitoring API Requests", monitAPI))
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func monitoringCostComponent(displayName string, unit string, usageTier string, description string, monthlyQuantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            displayName,
		Unit:            unit,
		UnitMultiplier:  1,
		MonthlyQuantity: monthlyQuantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Service:       strPtr("Stackdriver Monitoring"),
			ProductFamily: strPtr("ApplicationServices"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", Value: &description},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(usageTier),
		},
	}
}
