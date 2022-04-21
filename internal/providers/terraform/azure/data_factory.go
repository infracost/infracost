package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getDataFactoryRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_data_factory",
		RFunc: newDataFactory,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func newDataFactory(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{"resource_group_name"})

	r := &azure.DataFactory{
		Address: d.Address,
		Region:  region,
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
