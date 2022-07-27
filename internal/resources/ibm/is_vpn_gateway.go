package ibm

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// IsVpnGateway struct represents a VPN gateway to an IBM Cloud VPC
//
// Catalog information: https://cloud.ibm.com/vpc-ext/provision/vpngateway
// Resource information: https://cloud.ibm.com/docs/vpc?topic=vpc-vpn-overview
// Pricing information: https://www.ibm.com/cloud/vpc/pricing
type IsVpnGateway struct {
	Address string
	Region  string

	ConnectionHours *float64 `infracost_usage:"connection_hours"`
	InstanceHours   *float64 `infracost_usage:"instance_hours"`
}

// IsVpnGatewayUsageSchema defines a list which represents the usage schema of IsVpnGateway.
var IsVpnGatewayUsageSchema = []*schema.UsageItem{
	{Key: "connection_hours", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "instance_hours", DefaultValue: 0, ValueType: schema.Float64},
}

// PopulateUsage parses the u schema.UsageData into the IsVpnGateway.
// It uses the `infracost_usage` struct tags to populate data into the IsVpnGateway.
func (r *IsVpnGateway) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *IsVpnGateway) connectionHoursCostComponent() *schema.CostComponent {
	var quantity *decimal.Decimal
	if r.ConnectionHours != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.ConnectionHours))
	}
	return &schema.CostComponent{
		Name:            fmt.Sprintf("VPN connection hours %s", r.Region),
		Unit:            "Hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Region),
			Service:    strPtr("is.vpn"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", ValueRegex: strPtr(fmt.Sprintf("/%s/i", "gen2-vpn"))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("CONNECTION_HOURS"),
		},
	}
}

func (r *IsVpnGateway) instanceHoursCostComponent() *schema.CostComponent {
	var quantity *decimal.Decimal
	if r.InstanceHours != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.InstanceHours))
	}
	return &schema.CostComponent{
		Name:            fmt.Sprintf("VPN instance hours %s", r.Region),
		Unit:            "Hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Region),
			Service:    strPtr("is.vpn"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", ValueRegex: strPtr(fmt.Sprintf("/%s/i", "gen2-vpn"))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("INSTANCE_HOURS"),
		},
	}
}

// BuildResource builds a schema.Resource from a valid IsVpnGateway struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *IsVpnGateway) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.connectionHoursCostComponent(),
		r.instanceHoursCostComponent(),
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    IsVpnGatewayUsageSchema,
		CostComponents: costComponents,
	}
}
