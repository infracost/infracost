package scaleway

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetInstanceServerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "scaleway_instance_server",
		RFunc: NewInstanceServer,
	}
}

func NewInstanceServer(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	instanceType := d.Get("type").String()
	instanceImage := d.Get("image").String()
	instanceZone := d.Get("zone").String()

	fmt.Printf("%#v\n\n", d)

	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			instanceServerCostComponent(instanceZone, instanceType, instanceImage),
		},
	}
}

func instanceServerCostComponent(instanceZone string, instanceType string, instanceImage string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           fmt.Sprintf("Instance usage (%s, %s)", instanceImage, instanceType),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("scaleway"),
			Region:        strPtr(instanceZone),
			Service:       strPtr("Instance"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "type", Value: strPtr(instanceType)},
			},
		},
	}
}
