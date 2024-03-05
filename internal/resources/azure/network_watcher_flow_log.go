package azure

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// NetworkWatcherFlowLog struct represents Azure Network Watcher Flow Log
//
// From the Azure Network Watcher pricing page, this resource supports the
// 'Network Logs Collected' and 'Traffic Analytics' pricing.
//
// Other Network Monitor prices are supported in other resources, as specified
// in the NetworkWatcher resource struct.
//
// Resource information: https://azure.microsoft.com/en-gb/pricing/details/network-watcher/#overview
// Pricing information: https://azure.microsoft.com/en-gb/pricing/details/network-watcher/#pricing
type NetworkWatcherFlowLog struct {
	Address                               string
	Region                                string
	TrafficAnalyticsEnabled               bool
	TrafficAnalyticsAcceleratedProcessing bool

	MonthlyLogsCollectedGB *float64 `infracost_usage:"monthly_logs_collected_gb"`
}

// CoreType returns the name of this resource type
func (r *NetworkWatcherFlowLog) CoreType() string {
	return "NetworkWatcherFlowLog"
}

// UsageSchema defines a list which represents the usage schema of NetworkWatcherFlowLog.
func (r *NetworkWatcherFlowLog) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_logs_collected_gb", ValueType: schema.Float64, DefaultValue: 0.0},
	}
}

// PopulateUsage parses the u schema.UsageData into the NetworkWatcherFlowLog.
// It uses the `infracost_usage` struct tags to populate data into the NetworkWatcherFlowLog.
func (r *NetworkWatcherFlowLog) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid NetworkWatcherFlowLog struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *NetworkWatcherFlowLog) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.networkLogsCollectedCostComponent(),
	}

	if r.TrafficAnalyticsEnabled {
		costComponents = append(costComponents, r.trafficAnalyticsDataProcessedCostComponent())
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *NetworkWatcherFlowLog) networkLogsCollectedCostComponent() *schema.CostComponent {
	freeQuantity := decimal.NewFromInt(5)

	var qty *decimal.Decimal
	if r.MonthlyLogsCollectedGB != nil {
		// 5 GB per Network Watcher are free
		qty = decimalPtr(decimal.NewFromFloat(*r.MonthlyLogsCollectedGB).Sub(freeQuantity))
		if qty.LessThan(decimal.Zero) {
			qty = decimalPtr(decimal.Zero)
		}
	}

	return &schema.CostComponent{
		Name:            "Network logs collected (over 5GB)",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Network Watcher"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", Value: strPtr("Standard Network Logs Collected")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(freeQuantity.String()),
		},
		UsageBased: true,
	}
}

func (r *NetworkWatcherFlowLog) trafficAnalyticsDataProcessedCostComponent() *schema.CostComponent {
	var qty *decimal.Decimal
	if r.MonthlyLogsCollectedGB != nil {
		qty = decimalPtr(decimal.NewFromFloat(*r.MonthlyLogsCollectedGB))
	}

	meterName := "Standard Traffic Analytics Processing"
	suffix := "(60 min interval)"
	if r.TrafficAnalyticsAcceleratedProcessing {
		meterName = "Standard Traffic Analytics Processing at 10-Minute Interval"
		suffix = "(10 min interval)"
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Traffic Analytics data processed %s", suffix),
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Network Watcher"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", Value: strPtr(meterName)},
			},
		},
		UsageBased: true,
	}
}
