package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getServiceBusNamespaceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_servicebus_namespace",
		CoreRFunc: newServiceBusNamespace,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func newServiceBusNamespace(d *schema.ResourceData) schema.CoreResource {
	region := d.Region
	return &azure.ServiceBusNamespace{
		Address:  d.Address,
		Region:   region,
		SKU:      d.Get("sku").String(),
		Capacity: d.Get("capacity").Int(),
	}
}
