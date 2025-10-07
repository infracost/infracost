package azure

import (
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// VPNGatewayConnection represents a VPN Gateway connection, which is a billable component
// of a S2S VPN gateway. See VPNGateway for more information.
//
// More resource information here: https://docs.microsoft.com/en-us/azure/virtual-wan/virtual-wan-about
// Pricing information here: https://azure.microsoft.com/en-us/pricing/details/virtual-wan/
type VPNGatewayConnection struct {
	// Address is the unique name of the resource in the IAC language.
	Address string
	// Region is the azure region the VPNGatewayConnection is provisioned within.
	Region string
}

func (r *VPNGatewayConnection) CoreType() string {
	return "VPNGatewayConnection"
}

func (r *VPNGatewayConnection) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

// PopulateUsage parses the u schema.UsageData into the VPNGatewayConnection.
// It uses the `infracost_usage` struct tags to populate data into the VPNGatewayConnection.
func (r *VPNGatewayConnection) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid VPNGatewayConnection.
// It returns a VPNGatewayConnection as a schema.Resource with a single cost component representing the
// connection unit. The hourly quantity is set to 1 as VPNGatewayConnection represents a single connection unit.
//
// This method is called after the resource is initialised by an iac provider.
// See providers folder for more information.
func (r *VPNGatewayConnection) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name:        r.Address,
		UsageSchema: r.UsageSchema(),
		CostComponents: []*schema.CostComponent{
			{
				Name:           "S2S Connections",
				Unit:           "hours",
				UnitMultiplier: decimal.NewFromInt(1),
				HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("azure"),
					Region:        strPtr(r.Region),
					Service:       strPtr("Virtual WAN"),
					ProductFamily: strPtr("Networking"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "skuName", Value: strPtr("VPN S2S Connection Unit")},
					},
				},
				PriceFilter: &schema.PriceFilter{
					PurchaseOption: strPtr("Consumption"),
				},
			},
		},
	}
}
