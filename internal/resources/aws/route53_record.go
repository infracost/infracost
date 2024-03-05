package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/usage"
)

type Route53Record struct {
	Address                    string
	IsAlias                    bool
	MonthlyLatencyBasedQueries *int64 `infracost_usage:"monthly_latency_based_queries"`
	MonthlyGeoQueries          *int64 `infracost_usage:"monthly_geo_queries"`
	MonthlyStandardQueries     *int64 `infracost_usage:"monthly_standard_queries"`
}

func (r *Route53Record) CoreType() string {
	return "Route53Record"
}

func (r *Route53Record) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_latency_based_queries", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_geo_queries", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_standard_queries", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *Route53Record) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *Route53Record) BuildResource() *schema.Resource {
	if r.IsAlias {
		return &schema.Resource{
			Name:        r.Address,
			NoPrice:     true,
			IsSkipped:   true,
			UsageSchema: r.UsageSchema(),
		}
	}

	costComponents := []*schema.CostComponent{}
	limits := []int{1000000000}

	var numbOfStdQueries *decimal.Decimal
	if r.MonthlyStandardQueries != nil {
		numbOfStdQueries = decimalPtr(decimal.NewFromInt(*r.MonthlyStandardQueries))
		stdQueriesTiers := usage.CalculateTierBuckets(*numbOfStdQueries, limits)

		if stdQueriesTiers[0].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, queriesCostComponent("Standard queries (first 1B)", "DNS-Queries", "0", &stdQueriesTiers[0]))
		}

		if stdQueriesTiers[1].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, queriesCostComponent("Standard queries (over 1B)", "DNS-Queries", "1000000000", &stdQueriesTiers[1]))
		}
	} else {
		var unknown *decimal.Decimal

		costComponents = append(costComponents, queriesCostComponent("Standard queries (first 1B)", "DNS-Queries", "0", unknown))
	}

	var numbOfLBRQueries *decimal.Decimal
	if r.MonthlyLatencyBasedQueries != nil {
		numbOfLBRQueries = decimalPtr(decimal.NewFromInt(*r.MonthlyLatencyBasedQueries))
		lbrQueriesTiers := usage.CalculateTierBuckets(*numbOfLBRQueries, limits)

		if lbrQueriesTiers[0].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, queriesCostComponent("Latency based routing queries (first 1B)", "LBR-Queries", "0", &lbrQueriesTiers[0]))
		}

		if lbrQueriesTiers[1].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, queriesCostComponent("Latency based routing queries (over 1B)", "LBR-Queries", "1000000000", &lbrQueriesTiers[1]))
		}
	} else {
		var unknown *decimal.Decimal

		costComponents = append(costComponents, queriesCostComponent("Latency based routing queries (first 1B)", "LBR-Queries", "0", unknown))
	}

	var numbOfGeoQueries *decimal.Decimal
	if r.MonthlyGeoQueries != nil {
		numbOfGeoQueries = decimalPtr(decimal.NewFromInt(*r.MonthlyGeoQueries))
		geoQueriesTiers := usage.CalculateTierBuckets(*numbOfGeoQueries, limits)

		if geoQueriesTiers[0].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, queriesCostComponent("Geo DNS queries (first 1B)", "Geo-Queries", "0", &geoQueriesTiers[0]))
		}

		if geoQueriesTiers[1].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, queriesCostComponent("Geo DNS queries (over 1B)", "Geo-Queries", "1000000000", &geoQueriesTiers[1]))
		}
	} else {
		var unknown *decimal.Decimal

		costComponents = append(costComponents, queriesCostComponent("Geo DNS queries (first 1B)", "Geo-Queries", "0", unknown))
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func queriesCostComponent(displayName string, usageType string, usageTier string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            displayName,
		Unit:            "1M queries",
		UnitMultiplier:  decimal.NewFromInt(1000000),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Service:       strPtr("AmazonRoute53"),
			ProductFamily: strPtr("DNS Query"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", Value: &usageType},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(usageTier),
		},
		UsageBased: true,
	}
}
