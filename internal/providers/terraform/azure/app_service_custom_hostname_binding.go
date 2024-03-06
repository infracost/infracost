package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAppServiceCustomHostnameBindingRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_app_service_custom_hostname_binding",
		CoreRFunc: NewAppServiceCustomHostnameBinding,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}
func NewAppServiceCustomHostnameBinding(d *schema.ResourceData) schema.CoreResource {
	r := &azure.AppServiceCustomHostnameBinding{Address: d.Address, SSLState: d.Get("ssl_state").String()}
	r.Region = "Global"
	group := d.References("resource_group_name")
	if len(group) > 0 {
		r.Region = group[0].Get("location").String()
	}
	return r
}
