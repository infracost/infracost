package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getManagedRedisRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_managed_redis",
		CoreRFunc: newManagedRedis,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func newManagedRedis(d *schema.ResourceData) schema.CoreResource {
	return &azure.ManagedRedis{
		Address: d.Address,
		Region:  d.Region,
		SKU:     d.Get("sku_name").String(),
	}
}
