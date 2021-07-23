package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"
)

func GetRoute53RecordRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_route53_record",
		RFunc:               NewRoute53Record,
		ReferenceAttributes: []string{"alias.0.name"},
	}
}

func NewRoute53Record(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	if len(d.References("alias.0.name")) > 0 && d.References("alias.0.name")[0].Type != "aws_route53_record" {
		return &schema.Resource{
			NoPrice:   true,
			IsSkipped: true,
		}
	}

	costComponents := []*schema.CostComponent{}
	limits := []int{1000000000}

	var numbOfStdQueries *decimal.Decimal
	if u != nil && u.Get("monthly_standard_queries").Exists() {
		numbOfStdQueries = decimalPtr(decimal.NewFromInt(u.Get("monthly_standard_queries").Int()))
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
	if u != nil && u.Get("monthly_latency_based_queries").Exists() {
		numbOfLBRQueries = decimalPtr(decimal.NewFromInt(u.Get("monthly_latency_based_queries").Int()))
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
	if u != nil && u.Get("monthly_geo_queries").Exists() {
		numbOfGeoQueries = decimalPtr(decimal.NewFromInt(u.Get("monthly_geo_queries").Int()))
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
		Name:           d.Address,
		CostComponents: costComponents,
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
	}
}
