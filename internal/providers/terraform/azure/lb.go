package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getLoadBalancerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_lb",
		CoreRFunc: NewLB,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func NewLB(d *schema.ResourceData) schema.CoreResource {
	r := &azure.LB{
		Address: d.Address,
		Region:  d.Region,
		SKU:     d.Get("sku").String(),
	}
	return r
}
