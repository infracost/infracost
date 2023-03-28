package azure

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// PowerBIEmbedded struct represents a Power BI Embedded resource.
//
// Resource information: https://learn.microsoft.com/en-us/power-bi/developer/embedded/
// Pricing information: https://azure.microsoft.com/en-gb/pricing/details/power-bi-embedded/
type PowerBIEmbedded struct {
	Address string
	SKU     string
	Region  string
}

func (r *PowerBIEmbedded) CoreType() string {
	return "PowerBIEmbedded"
}

func (r *PowerBIEmbedded) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

// PopulateUsage parses the u schema.UsageData into the PowerBIEmbedded.
// It uses the `infracost_usage` struct tags to populate data into the PowerBIEmbedded.
func (r *PowerBIEmbedded) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid PowerBIEmbedded struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *PowerBIEmbedded) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name:           r.Address,
		CostComponents: []*schema.CostComponent{r.instanceUsageCostComponent()},
	}
}

func (r *PowerBIEmbedded) instanceUsageCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           fmt.Sprintf("Node usage (%s)", r.SKU),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Power BI Embedded"),
			ProductFamily: strPtr("Analytics"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr(r.SKU)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
