package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type ComputeVpnTunnel struct {
	Address *string
	Region  *string
}

var ComputeVpnTunnelUsageSchema = []*schema.UsageItem{}

func (r *ComputeVpnTunnel) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ComputeVpnTunnel) BuildResource() *schema.Resource {
	region := *r.Region
	return &schema.Resource{
		Name: *r.Address,
		CostComponents: []*schema.CostComponent{
			VPNTunnelInstance(region),
		}, UsageSchema: ComputeVpnTunnelUsageSchema,
	}
}

func VPNTunnelInstance(region string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "VPN Tunnel",
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(region),
			Service:       strPtr("Compute Engine"),
			ProductFamily: strPtr("Network"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "resourceGroup", Value: strPtr("VPNTunnel")},
			},
		},
	}
}
