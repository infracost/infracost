package scaleway

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetInstanceIPRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "scaleway_instance_ip",
		RFunc: NewInstanceIP,
	}
}

func NewInstanceIP(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	instanceZone := d.Get("zone").String()

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			instanceIPCostComponent(instanceZone),
		},
	}
}

func instanceIPCostComponent(instanceZone string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           "IP usage",
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("scaleway"),
			Region:        strPtr(instanceZone),
			Service:       strPtr("Flexible IP"),
			ProductFamily: strPtr("Compute"),
		},
	}
}
