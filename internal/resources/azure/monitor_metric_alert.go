package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// MonitorMetricAlert struct represents an Azure Monitor Metric Group.
//
// Resource information: https://learn.microsoft.com/en-us/azure/azure-monitor/alerts/alerts-overview
// Pricing information: https://azure.microsoft.com/en-us/pricing/details/monitor/
type MonitorMetricAlert struct {
	Address string
	Region  string

	Enabled                        bool
	ScopeCount                     int
	CriteriaDimensionsCount        int
	DynamicCriteriaDimensionsCount int
}

func (r *MonitorMetricAlert) CoreType() string {
	return "MonitorMetricAlert"
}

func (r *MonitorMetricAlert) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

// PopulateUsage parses the u schema.UsageData
// It uses the `infracost_usage` struct tags to populate data.
func (r *MonitorMetricAlert) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from the struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *MonitorMetricAlert) BuildResource() *schema.Resource {
	if !r.Enabled {
		return &schema.Resource{
			Name: r.Address,
		}
	}

	costComponents := []*schema.CostComponent{}

	totalTimeSeries := int64(r.ScopeCount * (r.CriteriaDimensionsCount + r.DynamicCriteriaDimensionsCount))
	if totalTimeSeries > 0 {
		costComponents = append(costComponents, r.alertMetricMonitoringCostComponent(totalTimeSeries))
	}

	dynamicTimeSeries := int64(r.ScopeCount * r.DynamicCriteriaDimensionsCount)
	if dynamicTimeSeries > 0 {
		costComponents = append(costComponents, r.dynamicThresholdCostComponent(dynamicTimeSeries))
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
	}
}

func (r *MonitorMetricAlert) alertMetricMonitoringCostComponent(quantity int64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Metrics monitoring",
		Unit:            "time-series",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(quantity)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        r.normalizedRegion(),
			Service:       strPtr("Azure Monitor"),
			ProductFamily: strPtr("Management and Governance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", Value: strPtr("Alerts Metric Monitored")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("10"),
		},
	}
}

func (r *MonitorMetricAlert) dynamicThresholdCostComponent(quantity int64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Dynamic threshold monitoring",
		Unit:            "time-series",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(quantity)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        r.normalizedRegion(),
			Service:       strPtr("Azure Monitor"),
			ProductFamily: strPtr("Management and Governance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", Value: strPtr("Alerts Dynamic Threshold")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("0"),
		},
	}
}

func (r *MonitorMetricAlert) normalizedRegion() *string {
	if r.Region == "global" {
		return strPtr("Global")
	}
	return strPtr(r.Region)
}
