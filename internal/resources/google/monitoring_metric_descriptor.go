package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"

	"github.com/dustin/go-humanize"

	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"
)

type MonitoringMetricDescriptor struct {
	Address                 *string
	MonthlyMonitoringDataMb *int64 `infracost_usage:"monthly_monitoring_data_mb"`
	MonthlyAPICalls         *int64 `infracost_usage:"monthly_api_calls"`
}

var MonitoringMetricDescriptorUsageSchema = []*schema.UsageItem{{Key: "monthly_monitoring_data_mb", ValueType: schema.Int64, DefaultValue: 0}, {Key: "monthly_api_calls", ValueType: schema.Int64, DefaultValue: 0}}

func (r *MonitoringMetricDescriptor) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *MonitoringMetricDescriptor) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}

	var monitData *decimal.Decimal
	if r.MonthlyMonitoringDataMb != nil {
		monitData = decimalPtr(decimal.NewFromInt(*r.MonthlyMonitoringDataMb))

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
	if r.MonthlyAPICalls != nil {
		monitAPI = decimalPtr(decimal.NewFromInt(*r.MonthlyAPICalls))
	}

	costComponents = append(costComponents, monitoringCostComponent("API calls", "calls", 1000, "1000000", "Monitoring API Requests", monitAPI))
	return &schema.Resource{
		Name:           *r.Address,
		CostComponents: costComponents, UsageSchema: MonitoringMetricDescriptorUsageSchema,
	}
}

func monitoringCostComponent(displayName string, unit string, unitMultiplier int, usageTier string, description string, monthlyQuantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            displayName,
		Unit:            humanize.SI(float64(unitMultiplier), "") + " " + unit,
		UnitMultiplier:  decimal.NewFromInt(int64(unitMultiplier)),
		MonthlyQuantity: monthlyQuantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Service:       strPtr("Stackdriver Monitoring"),
			ProductFamily: strPtr("ApplicationServices"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: strPtr(fmt.Sprintf("/%s/i", description))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(usageTier),
		},
	}
}
