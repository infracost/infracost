package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/usage"
)

type MonitoringMetricDescriptor struct {
	Address                 string
	MonthlyMonitoringDataMB *int64 `infracost_usage:"monthly_monitoring_data_mb"`
	MonthlyAPICalls         *int64 `infracost_usage:"monthly_api_calls"`
}

func (r *MonitoringMetricDescriptor) CoreType() string {
	return "MonitoringMetricDescriptor"
}

func (r *MonitoringMetricDescriptor) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_monitoring_data_mb", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_api_calls", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *MonitoringMetricDescriptor) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *MonitoringMetricDescriptor) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}

	var monitoringDataMB *decimal.Decimal
	if r.MonthlyMonitoringDataMB != nil {
		monitoringDataMB = decimalPtr(decimal.NewFromInt(*r.MonthlyMonitoringDataMB))

		monitoringDataTiers := usage.CalculateTierBuckets(*monitoringDataMB, []int{100000, 150000})

		if monitoringDataTiers[0].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, r.monitoringDataCostComponent("Monitoring data (first 100K)", "150", &monitoringDataTiers[0]))
		}

		if monitoringDataTiers[1].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, r.monitoringDataCostComponent("Monitoring data (next 150K)", "100000", &monitoringDataTiers[1]))
		}

		if monitoringDataTiers[2].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, r.monitoringDataCostComponent("Monitoring data (over 250K)", "250000", &monitoringDataTiers[2]))
		}
	} else {
		var unknown *decimal.Decimal
		costComponents = append(costComponents, r.monitoringDataCostComponent("Monitoring data (first 100K)", "150", unknown))
	}

	var apiCalls *decimal.Decimal
	if r.MonthlyAPICalls != nil {
		apiCalls = decimalPtr(decimal.NewFromInt(*r.MonthlyAPICalls))
	}

	costComponents = append(costComponents, r.apiCallsCostComponent(apiCalls))
	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *MonitoringMetricDescriptor) monitoringDataCostComponent(name string, usageTier string, monitoringDataMB *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "MB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: monitoringDataMB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Service:       strPtr("Cloud Monitoring"),
			ProductFamily: strPtr("ApplicationServices"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: strPtr("/Metric Volume/i")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(usageTier),
		},
		UsageBased: true,
	}
}

func (r *MonitoringMetricDescriptor) apiCallsCostComponent(apiCalls *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "API calls",
		Unit:            "1k calls",
		UnitMultiplier:  decimal.NewFromInt(1000),
		MonthlyQuantity: apiCalls,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Service:       strPtr("Cloud Monitoring"),
			ProductFamily: strPtr("ApplicationServices"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: strPtr("/Monitoring API Requests/i")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("1000000"),
		},
		UsageBased: true,
	}
}
