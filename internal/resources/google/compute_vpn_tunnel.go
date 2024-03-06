package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type ComputeVPNTunnel struct {
	Address string
	Region  string
}

func (r *ComputeVPNTunnel) CoreType() string {
	return "ComputeVPNTunnel"
}

func (r *ComputeVPNTunnel) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *ComputeVPNTunnel) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ComputeVPNTunnel) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			r.vpnTunnelCostComponent(),
		}, UsageSchema: r.UsageSchema(),
	}
}

func (r *ComputeVPNTunnel) vpnTunnelCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "VPN Tunnel",
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Compute Engine"),
			ProductFamily: strPtr("Network"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "resourceGroup", Value: strPtr("VPNTunnel")},
			},
		},
	}
}
