package azure

import (
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// ExpressRouteConnection represents an Express Route Gateway connection, which is a billable component
// of ExpressRouteGateway. See ExpressRouteGateway for more information.
//
// More resource information here: https://docs.microsoft.com/en-us/azure/virtual-wan/virtual-wan-about
// Pricing information here: https://azure.microsoft.com/en-us/pricing/details/virtual-wan/
type ExpressRouteConnection struct {
	// Address is the unique name of the resource in the IAC language.
	Address string
	// Region is the azure region the ExpressRouteConnection is provisioned within.
	Region string
}

func (e *ExpressRouteConnection) CoreType() string {
	return "ExpressRouteConnection"
}

// UsageSchema defines a list which represents the usage schema of EventGridTopic.
func (e *ExpressRouteConnection) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

// PopulateUsage parses the u schema.UsageData into the ExpressRouteConnection.
// It uses the `infracost_usage` struct tags to populate data into the ExpressRouteConnection.
func (e *ExpressRouteConnection) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(e, u)
}

// BuildResource builds a schema.Resource from a valid ExpressRouteConnection.
// It returns a ExpressRouteConnection as a schema.Resource with a single cost component representing the
// connection unit. The hourly quantity is set to 1 as ExpressRouteConnection represents a single connection unit.
//
// This method is called after the resource is initialised by an iac provider.
// See providers folder for more information.
func (e *ExpressRouteConnection) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name:        e.Address,
		UsageSchema: e.UsageSchema(),
		CostComponents: []*schema.CostComponent{
			{
				Name:           "ER Connections",
				Unit:           "hours",
				UnitMultiplier: decimal.NewFromInt(1),
				HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("azure"),
					Region:        strPtr(e.Region),
					Service:       strPtr("Virtual WAN"),
					ProductFamily: strPtr("Networking"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "skuName", Value: strPtr("ExpressRoute Connection Unit")},
					},
				},
				PriceFilter: &schema.PriceFilter{
					PurchaseOption: strPtr("Consumption"),
				},
			},
		},
	}
}
