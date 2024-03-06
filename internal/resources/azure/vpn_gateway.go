package azure

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// VPNGateway represents a Virtual WAN VPN gateway. It can represent a
// Point-to-site gateway (P2S) or a Site-to-site (S2S) gateway.
// Both gateways have similar price components on azure: Scale Unit & Connection Unit.
// However, S2S gateway connection costs are found through VPNGatewayConnection resource.
// Whereas P2S defines a usage param which is parsed below.
//
// More resource information here: https://docs.microsoft.com/en-us/azure/virtual-wan/virtual-wan-about
// Pricing information here: https://azure.microsoft.com/en-us/pricing/details/virtual-wan/
type VPNGateway struct {
	// Address is the unique name of the resource in the IAC language.
	Address string
	// Region is the azure region the VPNGateway is provisioned within.
	Region string
	// ScaleUnits represents a unit defined to pick an aggregate throughput of a gateway in Virtual hub.
	// 1 scale unit of VPN = 500 Mbps.
	ScaleUnits int64
	// Type represents the type of WAN Vpn Gateway, it can be one of: P2S|S2S.
	Type string

	// MonthlyP2SConnectionHrs represents a usage cost for the number of connection hours that the vpn
	// gateway has been in use for. Can be a fraction to denote smaller time increments lower than a whole hour.
	// This usage cost is only applicable for point to site vpns.
	MonthlyP2SConnectionHrs *float64 `infracost_usage:"monthly_p2s_connections_hrs"`
}

func (v *VPNGateway) CoreType() string {
	return "VPNGateway"
}

// UsageSchema defines a list which represents the usage schema of VPNGateway if of type P2S.
func (v *VPNGateway) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_p2s_connections_hrs", DefaultValue: 0, ValueType: schema.Float64},
	}
}

// PopulateUsage parses the u schema.UsageData into the VPNGateway.
// It uses the `infracost_usage` struct tags to populate data into the VPNGateway.
func (v *VPNGateway) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(v, u)
}

// BuildResource builds a schema.Resource from a valid VPNGateway. It returns different Resources
// based on the VPNGateway.Type. If type Point to Site (P2S) it will include a usage cost component
// based on the connection usage. For other cases (S2S) it will just include a single scale unit
// cost component. See VPNGatewayConnection for S2S connection costs associated with S2S gateway.
//
// This method is called after the resource is initialised by an iac provider.
// See providers folder for more information.
func (v *VPNGateway) BuildResource() *schema.Resource {
	if v.Type == "P2S" {
		return v.buildP2SResource()
	}

	return v.buildS2SResource()
}

func (v *VPNGateway) buildS2SResource() *schema.Resource {
	return &schema.Resource{
		Name:           v.Address,
		CostComponents: []*schema.CostComponent{v.scaleUnitComponent()},
	}
}

func (v *VPNGateway) buildP2SResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		v.scaleUnitComponent(),
		v.connectionUnitComponent(),
	}

	return &schema.Resource{
		Name:           v.Address,
		CostComponents: costComponents,
		UsageSchema:    v.UsageSchema(),
	}
}

func (v *VPNGateway) scaleUnitComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           fmt.Sprintf("%s scale units (500 Mbps)", v.Type),
		Unit:           "scale units",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(v.ScaleUnits)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(v.Region),
			Service:       strPtr("Virtual WAN"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr(fmt.Sprintf("VPN %s Scale Unit", v.Type))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func (v *VPNGateway) connectionUnitComponent() *schema.CostComponent {
	var connections float64
	if v.MonthlyP2SConnectionHrs != nil {
		connections = *v.MonthlyP2SConnectionHrs
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("%s connections", v.Type),
		Unit:            "hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromFloat(connections)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(v.Region),
			Service:       strPtr("Virtual WAN"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr(fmt.Sprintf("VPN %s Connection Unit", v.Type))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
		UsageBased: true,
	}

}
