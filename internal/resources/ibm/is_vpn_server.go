package ibm

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// IsVpnServer struct represents a VPN server for IBM Cloud VPC
//
// Catalog information: https://cloud.ibm.com/vpc-ext/provision/vpnserver
// Resource information: https://cloud.ibm.com/docs/vpc?topic=vpc-vpn-overview#client-to-site-vpn-server
// Pricing information: https://www.ibm.com/cloud/vpc/pricing
type IsVpnServer struct {
	Address                string
	Region                 string
	MonthlyConnectionHours *float64 `infracost_usage:"is.vpn-server_CONNECTION_HOURS"`
	MonthlyInstanceHours   *float64 `infracost_usage:"is.vpn-server_INSTANCE_HOURS"`
}

// IsVpnServerUsageSchema defines a list which represents the usage schema of IsVpnServer.
var IsVpnServerUsageSchema = []*schema.UsageItem{
	{Key: "is.vpn-server_CONNECTION_HOURS", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "is.vpn-server_INSTANCE_HOURS", DefaultValue: 0, ValueType: schema.Float64},
}

// PopulateUsage parses the u schema.UsageData into the IsVpnServer.
// It uses the `infracost_usage` struct tags to populate data into the IsVpnServer.
func (r *IsVpnServer) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *IsVpnServer) connectionHoursCostComponent() *schema.CostComponent {
	var quantity *decimal.Decimal
	if r.MonthlyConnectionHours != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.MonthlyConnectionHours))
	}
	return &schema.CostComponent{
		Name:            fmt.Sprintf("VPN connection hours %s", r.Region),
		Unit:            "Hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Region),
			Service:    strPtr("is.vpn-server"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", ValueRegex: strPtr(fmt.Sprintf("/%s/i", "gen2-vpn-server"))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("CONNECTION_HOURS"),
		},
	}
}

func (r *IsVpnServer) instanceHoursCostComponent() *schema.CostComponent {
	var quantity *decimal.Decimal
	if r.MonthlyInstanceHours != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.MonthlyInstanceHours))
	}
	return &schema.CostComponent{
		Name:            fmt.Sprintf("VPN instance hours %s", r.Region),
		Unit:            "Hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Region),
			Service:    strPtr("is.vpn-server"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", ValueRegex: strPtr(fmt.Sprintf("/%s/i", "gen2-vpn-server"))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("INSTANCE_HOURS"),
		},
	}
}

// BuildResource builds a schema.Resource from a valid IsVpnServer struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *IsVpnServer) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.connectionHoursCostComponent(),
		r.instanceHoursCostComponent(),
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    IsVpnServerUsageSchema,
		CostComponents: costComponents,
	}
}
