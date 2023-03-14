package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func getLinuxFunctionAppRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name: "azurerm_linux_function_app",
		ReferenceAttributes: []string{
			"service_plan_id",
		},
		CoreRFunc: func(d *schema.ResourceData) schema.CoreResource {
			return newFunctionApp(d)
		},
	}
}
