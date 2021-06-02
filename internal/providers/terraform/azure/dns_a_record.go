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
	group := d.References("resource_group_name")[0]

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: dnsQueriesCostComponent(d, u, group),
	}
}
func dnsQueriesCostComponent(d *schema.ResourceData, u *schema.UsageData, group *schema.ResourceData) []*schema.CostComponent {
	var monthlyQueries *decimal.Decimal
	var requestQuantities []decimal.Decimal
	costComponents := make([]*schema.CostComponent, 0)
	requests := []int{1000000000}

	location := group.Get("location").String()

	if strings.HasPrefix(strings.ToLower(location), "usgov") {
		location = "US Gov Zone 1"
	}
	if strings.HasPrefix(strings.ToLower(location), "germany") {
		location = "DE Zone 1"
	}
	if strings.HasPrefix(strings.ToLower(location), "china") {
		location = "Zone 1 (China)"
	}
	if location != "US Gov Zone 1" && location != "DE Zone 1" && location != "Zone 1 (China)" {
		location = "Zone 1"
	}

	if u != nil && u.Get("monthly_queries").Exists() {
		monthlyQueries = decimalPtr(decimal.NewFromInt(u.Get("monthly_queries").Int()))
		requestQuantities = usage.CalculateTierBuckets(*monthlyQueries, requests)
		firstBqueries := requestQuantities[0].Div(decimal.NewFromInt(1000000))
		overBqueries := requestQuantities[1].Div(decimal.NewFromInt(1000000))
		costComponents = append(costComponents, dnsQueriesFirstCostComponent(location, "DNS queries (first 1B)", "0", &firstBqueries))

		if requestQuantities[1].GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, dnsQueriesFirstCostComponent(location, "DNS queries (over 1B)", "1000", &overBqueries))
		}
	} else {
		var unknown *decimal.Decimal

		costComponents = append(costComponents, dnsQueriesFirstCostComponent(location, "DNS queries (first 1B)", "0", unknown))
	}

	return costComponents
}

func dnsQueriesFirstCostComponent(location, name, startUsage string, monthlyQueries *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "1M queries",
		UnitMultiplier:  1,
		MonthlyQuantity: monthlyQueries,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(location),
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
