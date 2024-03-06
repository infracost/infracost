package azure

import (
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// ExpressRouteGateway is a Virtual WAN gateway that provides direct connectivity to Azure cloud services.
// All transferred data is not encrypted, and do not go over the public Internet.
//
// More resource information here: https://docs.microsoft.com/en-us/azure/expressroute/expressroute-about-virtual-network-gateways
// Pricing information here: https://azure.microsoft.com/en-us/pricing/details/virtual-wan/
type ExpressRouteGateway struct {
	// Address is the unique name of the resource in the IAC language.
	Address string
	// Region is the azure region the VPNGateway is provisioned within.
	Region string
	// ScaleUnits represents a unit defined to pick an aggregate throughput of a gateway in Virtual hub.
	// 1 scale unit of ExpressRoute = 2 Gbps.
	ScaleUnits int64
}

func (e *ExpressRouteGateway) CoreType() string {
	return "ExpressRouteGateway"
}

// UsageSchema defines a list which represents the usage schema of EventGridTopic.
func (e *ExpressRouteGateway) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

// PopulateUsage parses the u schema.UsageData into the ExpressRouteGateway.
// It uses the `infracost_usage` struct tags to populate data into the ExpressRouteGateway.
func (e *ExpressRouteGateway) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(e, u)
}

// BuildResource builds a schema.Resource from a valid ExpressRouteGateway.
// It returns ExpressRouteGateway with a single cost component "ER scale units".
// See more about scale units reading ExpressRouteGateway.ScaleUnits.
//
// This method is called after the resource is initialised by an iac provider.
// See providers folder for more information.
func (e *ExpressRouteGateway) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name:        e.Address,
		UsageSchema: e.UsageSchema(),
		CostComponents: []*schema.CostComponent{
			{
				Name:           "ER scale units (2 Gbps)",
				Unit:           "scale units",
				UnitMultiplier: schema.HourToMonthUnitMultiplier,
				HourlyQuantity: decimalPtr(decimal.NewFromInt(e.ScaleUnits)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("azure"),
					Region:        strPtr(e.Region),
					Service:       strPtr("Virtual WAN"),
					ProductFamily: strPtr("Networking"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "skuName", Value: strPtr("ExpressRoute Scale Unit")},
					},
				},
				PriceFilter: &schema.PriceFilter{
					PurchaseOption: strPtr("Consumption"),
				},
			},
		},
	}
}
