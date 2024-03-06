package azure

import (
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// NetworkWatcher struct represents Azure Network Watcher.
//
// From the Azure Network Watcher pricing page, this resource supports the
// 'Network Diagnostic Checks' pricing.
//
// The other prices are supported as follows:
//
//   - 'Network Logs Collected' and 'Traffic Analytics' are counted against the
//     azurerm_network_watcher_flow_log resource.
//
//   - 'Connection Monitor' is counted against the
//     azurerm_network_connection_monitor resource.
//
//   - 'Network Performance Monitor' charges are not supported since they are
//     deprecated and do not have an equivalent resource.
//
// Resource information: https://azure.microsoft.com/en-gb/pricing/details/network-watcher/#overview
// Pricing information: https://azure.microsoft.com/en-gb/pricing/details/network-watcher/#pricing
type NetworkWatcher struct {
	Address string
	Region  string

	MonthlyDiagnosticChecks *int64 `infracost_usage:"monthly_diagnostic_checks"`
}

// CoreType returns the name of this resource type
func (r *NetworkWatcher) CoreType() string {
	return "NetworkWatcher"
}

// UsageSchema defines a list which represents the usage schema of NetworkWatcher.
func (r *NetworkWatcher) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_diagnostic_checks", DefaultValue: 0, ValueType: schema.Int64},
	}
}

// PopulateUsage parses the u schema.UsageData into the NetworkWatcher.
// It uses the `infracost_usage` struct tags to populate data into the NetworkWatcher.
func (r *NetworkWatcher) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid NetworkWatcher struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *NetworkWatcher) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.networkDiagnosticToolCostComponent(),
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *NetworkWatcher) networkDiagnosticToolCostComponent() *schema.CostComponent {
	freeQuantity := decimal.NewFromInt(1000)

	var qty *decimal.Decimal
	if r.MonthlyDiagnosticChecks != nil {
		// 1000 checks per Network Watcher are free
		qty = decimalPtr(decimal.NewFromInt(*r.MonthlyDiagnosticChecks).Sub(freeQuantity))
		if qty.LessThan(decimal.Zero) {
			qty = decimalPtr(decimal.Zero)
		}
	}

	return &schema.CostComponent{
		Name:            "Network diagnostic tool (over 1,000 checks)",
		Unit:            "checks",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Network Watcher"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", Value: strPtr("Standard Diagnostic Tool API")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(freeQuantity.String()),
		},
		UsageBased: true,
	}
}
