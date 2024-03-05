package aws

import (
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// NetworkfirewallFirewall struct represents an AWS Network Firewall Firewall resource.
//
// Resource information: https://aws.amazon.com/network-firewall/
// Pricing information: https://aws.amazon.com/network-firewall/pricing/
type NetworkfirewallFirewall struct {
	Address string
	Region  string

	MonthlyDataProcessedGB *float64 `infracost_usage:"monthly_data_processed_gb"`
}

// NetworkfirewallFirewallUsageSchema defines a list which represents the usage schema of NetworkfirewallFirewall.
func (r *NetworkfirewallFirewall) CoreType() string {
	return "NetworkfirewallFirewall"
}

func (r *NetworkfirewallFirewall) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_data_processed_gb", DefaultValue: 0, ValueType: schema.Float64},
	}
}

// PopulateUsage parses the u schema.UsageData into the NetworkfirewallFirewall.
// It uses the `infracost_usage` struct tags to populate data into the NetworkfirewallFirewall.
func (r *NetworkfirewallFirewall) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid NetworkfirewallFirewall struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *NetworkfirewallFirewall) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.endpointCostComponent(),
		r.dataProcessedCostComponent(),
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *NetworkfirewallFirewall) endpointCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "Network Firewall Endpoint",
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AWSNetworkFirewall"),
			ProductFamily: strPtr("AWS Firewall"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("^[A-Z0-9]*-Endpoint-Hour$")},
			},
		},
	}
}

func (r *NetworkfirewallFirewall) dataProcessedCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Data Processed",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: floatPtrToDecimalPtr(r.MonthlyDataProcessedGB),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AWSNetworkFirewall"),
			ProductFamily: strPtr("AWS Firewall"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: regexPtr("^[A-Z0-9]*-Traffic-GB-Processed$")},
			},
		},
		UsageBased: true,
	}
}
