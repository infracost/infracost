package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getDataFactoryRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_data_factory",
		CoreRFunc: newDataFactory,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func newDataFactory(d *schema.ResourceData) schema.CoreResource {
	region := d.Region

	r := &azure.DataFactory{
		Address: d.Address,
		Region:  region,
	}
	return r
}
