package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
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
		CostComponents: DNSqueriesCostComponent(d, u, group),
	}
}
func DNSqueriesFirstCostComponent(location string, monthlyQueries *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "DNS queries (first 1B)",
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
			StartUsageAmount: strPtr("0"),
		},
	}
}
func DNSqueriesOverCostComponent(location string, monthlyQueries *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "DNS queries (over 1B)",
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
			StartUsageAmount: strPtr("1000"),
		},
	}
}

func DNSqueriesCostComponent(d *schema.ResourceData, u *schema.UsageData, group *schema.ResourceData) []*schema.CostComponent {
	var monthlyQueries, firstOneBQueries *decimal.Decimal
	costComponents := make([]*schema.CostComponent, 0)

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

	if u != nil && u.Get("monthly_queries").Type != gjson.Null {
		firstOneBQueries = decimalPtr(decimal.NewFromInt(1000))
		monthlyQueries = decimalPtr(decimal.NewFromInt(u.Get("monthly_queries").Int()))

		if monthlyQueries.GreaterThan(*firstOneBQueries) {
			overOneBQueries := decimalPtr(decimal.NewFromInt(u.Get("monthly_queries").Int())).Sub(*firstOneBQueries)
			monthlyQueries = &overOneBQueries
			costComponents = append(costComponents, DNSqueriesOverCostComponent(location, monthlyQueries))
		}
	}

	costComponents = append(costComponents, DNSqueriesFirstCostComponent(location, firstOneBQueries))
	return costComponents
}
