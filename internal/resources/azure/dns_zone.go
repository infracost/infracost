package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"strings"

	"github.com/shopspring/decimal"
)

type DNSZone struct {
	Address string
	Region  string
}

func (r *DNSZone) CoreType() string {
	return "DNSZone"
}

func (r *DNSZone) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *DNSZone) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *DNSZone) BuildResource() *schema.Resource {

	var region string
	if strings.HasPrefix(strings.ToLower(r.Region), "usgov") {
		region = "US Gov Zone 1"
	} else if strings.HasPrefix(strings.ToLower(r.Region), "germany") {
		region = "DE Zone 1"
	} else if strings.HasPrefix(strings.ToLower(r.Region), "china") {
		region = "Zone 1 (China)"
	} else {
		region = "Zone 1"
	}

	costComponents := make([]*schema.CostComponent, 0)

	costComponents = append(costComponents, hostedPublicZoneCostComponent(region))
	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func hostedPublicZoneCostComponent(region string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Hosted zone",
		Unit:            "months",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Azure DNS"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "meterName", ValueRegex: regexPtr("Public Zone(s)?")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr("0"),
		},
	}
}
