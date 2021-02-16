package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"

	"github.com/shopspring/decimal"
)

func GetRoute53ResolverEndpointRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_route53_resolver_endpoint",
		RFunc: NewRoute53ResolverEndpoint,
	}
}

func NewRoute53ResolverEndpoint(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	resolverEndpointCount := int64(len(d.Get("ip_address").Array()))

	costComponents := []*schema.CostComponent{
		{
			Name:           "Resolver endpoints",
			Unit:           "hours",
			UnitMultiplier: 1,
			HourlyQuantity: decimalPtr(decimal.NewFromInt(resolverEndpointCount)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonRoute53"),
				ProductFamily: strPtr("DNS Query"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/ResolverNetworkInterface/")},
				},
			},
		},
	}

	queryTierLimits := []int{1000000000}

	if u != nil && u.Get("monthly_queries").Exists() {
		monthlyQueries := decimal.NewFromInt(u.Get("monthly_queries").Int())
		dnsQueriesTier := usage.CalculateTierBuckets(monthlyQueries, queryTierLimits)
		tierOne := dnsQueriesTier[0]
		tierTwo := dnsQueriesTier[1]

		if tierOne.GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, dnsQueriesCostComponent(region, "DNS queries (first 1B)", "0", &tierOne))
		}

		if tierTwo.GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, dnsQueriesCostComponent(region, "DNS queries (over 1B)", "1000000000", &tierTwo))
		}

	} else {
		var unknown *decimal.Decimal
		costComponents = append(costComponents, dnsQueriesCostComponent(region, "DNS queries (first 1B)", "0", unknown))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func dnsQueriesCostComponent(region string, displayName string, usageTier string, monthlyQueries *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            displayName,
		Unit:            "queries",
		UnitMultiplier:  1000000,
		MonthlyQuantity: monthlyQueries,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonRoute53"),
			ProductFamily: strPtr("DNS Query"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/DNS-Queries/")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(usageTier),
		},
	}
}
