package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"
)

func GetAzureRMDNSaRecordRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_dns_a_record",
		RFunc: NewAzureRMDNSaRecord,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func NewAzureRMDNSaRecord(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: dnsQueriesCostComponent(d, u),
	}
}
func dnsQueriesCostComponent(d *schema.ResourceData, u *schema.UsageData) []*schema.CostComponent {
	region := lookupRegion(d, []string{"resource_group_name"})

	var monthlyQueries *decimal.Decimal
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

	if u != nil && u.Get("monthly_queries").Exists() {
		monthlyQueries = decimalPtr(decimal.NewFromInt(u.Get("monthly_queries").Int()))
		requestQuantities = usage.CalculateTierBuckets(*monthlyQueries, requests)
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
	}
}
