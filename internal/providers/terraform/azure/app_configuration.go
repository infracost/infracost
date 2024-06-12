package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAppConfigurationRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_app_configuration",
		CoreRFunc: newAppConfiguration,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func newAppConfiguration(d *schema.ResourceData) schema.CoreResource {
	region := d.Region
	sku := strings.ToLower(strings.TrimSpace(d.Get("sku").String()))
	if sku == "" {
		sku = "free"
	}
	array := d.Get("replica").Array()
	return &azure.AppConfiguration{
		Address:  d.Address,
		Region:   region,
		SKU:      sku,
		Replicas: int64(len(array)),
	}
}
