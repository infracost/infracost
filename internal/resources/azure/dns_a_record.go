package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/usage"
)

type DNSARecord struct {
	Address        string
	Region         string
	MonthlyQueries *int64 `infracost_usage:"monthly_queries"`
}

func (r *DNSARecord) CoreType() string {
	return "DNSARecord"
}

func (r *DNSARecord) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{{Key: "monthly_queries", ValueType: schema.Int64, DefaultValue: 0}}
}

func (r *DNSARecord) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *DNSARecord) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name:           r.Address,
		CostComponents: dnsQueriesCostComponent(r.Region, r.MonthlyQueries),
		UsageSchema:    r.UsageSchema(),
	}
}
func dnsQueriesCostComponent(region string, monthlyQueries *int64) []*schema.CostComponent {
	var monthlyQueriesDec *decimal.Decimal
	var requestQuantities []decimal.Decimal
	costComponents := make([]*schema.CostComponent, 0)
	requests := []int{1000000000}

	if strings.HasPrefix(strings.ToLower(region), "usgov") {
		region = "US Gov Zone 1"
	}
	if strings.HasPrefix(strings.ToLower(region), "germany") {
		region = "DE Zone 1"
	}
	if strings.HasPrefix(strings.ToLower(region), "china") {
		region = "Zone 1 (China)"
	}
	if region != "US Gov Zone 1" && region != "DE Zone 1" && region != "Zone 1 (China)" {
		region = "Zone 1"
	}

	if monthlyQueries != nil {
		monthlyQueriesDec = decimalPtr(decimal.NewFromInt(*monthlyQueries))
		requestQuantities = usage.CalculateTierBuckets(*monthlyQueriesDec, requests)
		firstBqueries := requestQuantities[0].Div(decimal.NewFromInt(1000000))
		overBqueries := requestQuantities[1].Div(decimal.NewFromInt(1000000))
		costComponents = append(costComponents, dnsQueriesFirstCostComponent(region, "DNS queries (first 1B)", "0", &firstBqueries))

		if requestQuantities[1].GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, dnsQueriesFirstCostComponent(region, "DNS queries (over 1B)", "1000", &overBqueries))
		}
	} else {
		var unknown *decimal.Decimal

		costComponents = append(costComponents, dnsQueriesFirstCostComponent(region, "DNS queries (first 1B)", "0", unknown))
	}

	return costComponents
}

func dnsQueriesFirstCostComponent(region, name, startUsage string, monthlyQueries *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "1M queries",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: monthlyQueries,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Azure DNS"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", Value: strPtr("Public Queries")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: &startUsage,
		},
		UsageBased: true,
	}
}
