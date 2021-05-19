package google

import (
	"github.com/dustin/go-humanize"
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
	if u != nil && u.Get("monthly_monitoring_data_mb").Exists() {
		monitData = decimalPtr(decimal.NewFromInt(u.Get("monthly_monitoring_data_mb").Int()))

		apiCallsLimits := []int{100000, 150000}

		apiCallsTiers := usage.CalculateTierBuckets(*monitData, apiCallsLimits)

		if apiCallsTiers[0].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, monitoringCostComponent("Monitoring data (first 100K)", "MB", 1, "150", "Metric Volume", &apiCallsTiers[0]))
		}

		if apiCallsTiers[1].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, monitoringCostComponent("Monitoring data (next 150K)", "MB", 1, "100000", "Metric Volume", &apiCallsTiers[1]))
		}

		if apiCallsTiers[2].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, monitoringCostComponent("Monitoring data (over 250K)", "MB", 1, "250000", "Metric Volume", &apiCallsTiers[2]))
		}
	} else {
		var unknown *decimal.Decimal
		costComponents = append(costComponents, monitoringCostComponent("Monitoring data (first 100K)", "MB", 1, "150", "Metric Volume", unknown))
	}

	var monitAPI *decimal.Decimal
	if u != nil && u.Get("monthly_api_calls").Exists() {
		monitAPI = decimalPtr(decimal.NewFromInt(u.Get("monthly_api_calls").Int()))
	}

	costComponents = append(costComponents, monitoringCostComponent("API calls", "calls", 1000, "1000000", "Monitoring API Requests", monitAPI))
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func monitoringCostComponent(displayName string, unit string, unitMultiplier int, usageTier string, description string, monthlyQuantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            displayName,
		Unit:            humanize.SI(float64(unitMultiplier), "") + " " + unit,
		UnitMultiplier:  unitMultiplier,
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
