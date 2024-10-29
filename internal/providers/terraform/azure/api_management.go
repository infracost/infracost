package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAPIManagementRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_api_management",
		CoreRFunc: NewAPIManagement,
		ReferenceAttributes: []string{
			"certificate_id",
		},
	}
}
func NewAPIManagement(d *schema.ResourceData) schema.CoreResource {
	r := &azure.APIManagement{Address: d.Address, SKUName: d.Get("sku_name").String(), Region: d.Region}
	return r
}
