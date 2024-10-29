package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/infracost/infracost/internal/usage"

	"github.com/shopspring/decimal"
)

type Route53ResolverEndpoint struct {
	Address           string
	Region            string
	ResolverEndpoints int64
	MonthlyQueries    *int64 `infracost_usage:"monthly_queries"`
}

func (r *Route53ResolverEndpoint) CoreType() string {
	return "Route53ResolverEndpoint"
}

func (r *Route53ResolverEndpoint) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_queries", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *Route53ResolverEndpoint) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *Route53ResolverEndpoint) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		{
			Name:           "Resolver endpoints",
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: decimalPtr(decimal.NewFromInt(r.ResolverEndpoints)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(r.Region),
				Service:       strPtr("AmazonRoute53"),
				ProductFamily: strPtr("DNS Query"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/ResolverNetworkInterface$/")},
				},
			},
		},
	}

	queryTierLimits := []int{1000000000}

	if r.MonthlyQueries != nil {
		monthlyQueries := decimal.NewFromInt(*r.MonthlyQueries)
		dnsQueriesTier := usage.CalculateTierBuckets(monthlyQueries, queryTierLimits)
		tierOne := dnsQueriesTier[0]
		tierTwo := dnsQueriesTier[1]

		if tierOne.GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, r.queriesCostComponent("DNS queries (first 1B)", "0", &tierOne))
		}

		if tierTwo.GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, r.queriesCostComponent("DNS queries (over 1B)", "1000000000", &tierTwo))
		}

	} else {
		var unknown *decimal.Decimal
		costComponents = append(costComponents, r.queriesCostComponent("DNS queries (first 1B)", "0", unknown))
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *Route53ResolverEndpoint) queriesCostComponent(displayName string, usageTier string, monthlyQueries *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            displayName,
		Unit:            "1M queries",
		UnitMultiplier:  decimal.NewFromInt(1000000),
		MonthlyQuantity: monthlyQueries,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonRoute53"),
			ProductFamily: strPtr("DNS Query"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/DNS-Queries/")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(usageTier),
		},
		UsageBased: true,
	}
}
