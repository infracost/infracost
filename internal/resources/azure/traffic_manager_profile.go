package azure

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
)

// TrafficManagerProfile struct represents an Azure Traffic Manager profile.
//
// Resource information: https://learn.microsoft.com/en-us/azure/traffic-manager/traffic-manager-overview
// Pricing information: https://azure.microsoft.com/en-us/pricing/details/traffic-manager/#pricing
type TrafficManagerProfile struct {
	Address string
	Region  string

	Enabled            bool
	TrafficViewEnabled bool

	MonthlyDNSQueries            *int64 `infracost_usage:"monthly_dns_queries"`
	MonthlyTrafficViewDataPoints *int64 `infracost_usage:"monthly_traffic_view_data_points"`
}

// CoreType returns the name of this resource type
func (r *TrafficManagerProfile) CoreType() string {
	return "TrafficManagerProfile"
}

// UsageSchema defines a list which represents the usage schema of TrafficManagerProfile.
func (r *TrafficManagerProfile) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_dns_queries", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_traffic_view_data_points", DefaultValue: 0, ValueType: schema.Int64},
	}
}

// PopulateUsage parses the u schema.UsageData into the TrafficManagerProfile.
// It uses the `infracost_usage` struct tags to populate data into the TrafficManagerProfile.
func (r *TrafficManagerProfile) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid TrafficManagerProfile struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *TrafficManagerProfile) BuildResource() *schema.Resource {
	if !r.Enabled {
		return &schema.Resource{
			Name: r.Address,
		}
	}

	costComponents := []*schema.CostComponent{}

	if r.MonthlyDNSQueries != nil {
		dnsQuantities := usage.CalculateTierBuckets(
			*decimalPtr(decimal.NewFromInt(*r.MonthlyDNSQueries)),
			[]int{1000000000},
		)
		costComponents = append(costComponents, r.dnsQueriesCostComponent(&dnsQuantities[0], "first 1B", "0"))

		if dnsQuantities[1].GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, r.dnsQueriesCostComponent(&dnsQuantities[1], "over 1B", "1000"))
		}
	} else {
		costComponents = append(costComponents, r.dnsQueriesCostComponent(nil, "first 1B", "0"))
	}

	if r.TrafficViewEnabled {
		costComponents = append(costComponents, r.trafficViewCostComponent())
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *TrafficManagerProfile) dnsQueriesCostComponent(q *decimal.Decimal, tierName, startUsage string) *schema.CostComponent {
	var millions *decimal.Decimal
	if q != nil {
		millions = decimalPtr(q.Div(decimal.NewFromInt(1000000)))
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("DNS queries (%s)", tierName),
		Unit:            "1M queries",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: millions,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(trafficManagerBillingRegion(r.Region)),
			Service:       strPtr("Traffic Manager"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr("Azure Endpoint")},
				{Key: "meterName", Value: strPtr("DNS Queries")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(startUsage),
		},
		UsageBased: true,
	}
}

func (r *TrafficManagerProfile) trafficViewCostComponent() *schema.CostComponent {
	var millions *decimal.Decimal
	if r.MonthlyTrafficViewDataPoints != nil {
		millions = decimalPtr(decimal.NewFromInt(*r.MonthlyTrafficViewDataPoints).Div(decimal.NewFromInt(1000000)))
	}

	return &schema.CostComponent{
		Name:            "Traffic view",
		Unit:            "1M data points",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: millions,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(trafficManagerBillingRegion(r.Region)),
			Service:       strPtr("Traffic Manager"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr("Traffic View")},
				{Key: "meterName", Value: strPtr("Traffic View Data Points Processed")},
			},
		},
		UsageBased: true,
	}
}

func trafficManagerBillingRegion(region string) string {
	switch {
	case strings.Contains(strings.ToLower(region), "usgov"):
		return "US Gov"
	case strings.Contains(strings.ToLower(region), "china"):
		return "China"
	case strings.Contains(strings.ToLower(region), "germany"):
		return "Germany"
	default:
		return "Global"
	}
}
