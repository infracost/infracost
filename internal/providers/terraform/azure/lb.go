package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getLoadBalancerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_lb",
		RFunc: NewLB,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func NewLB(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.LB{
		Address: d.Address,
		Region: lookupRegion(d, []string{"resource_group_name"}),
		SKU: d.Get("sku").String(),
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
