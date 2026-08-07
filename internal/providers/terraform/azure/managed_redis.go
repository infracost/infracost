package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
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
	highAvailability := true
	if d.Get("high_availability_enabled").Type != gjson.Null && !d.Get("high_availability_enabled").Bool() {
		highAvailability = false
	}

	return &azure.ManagedRedis{
		Address:          d.Address,
		Region:           d.Region,
		SKU:              d.Get("sku_name").String(),
		HighAvailability: highAvailability,
	}
}
