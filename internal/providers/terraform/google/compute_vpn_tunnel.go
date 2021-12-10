package google

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetComputeVPNTunnelRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_compute_vpn_tunnel",
		RFunc: NewComputeVPNTunnel,
	}
}

func NewComputeVPNTunnel(ctx *config.ProjectContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			VPNTunnelInstance(region),
		},
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
