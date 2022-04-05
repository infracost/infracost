package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAzureRMAppServiceCertificateBindingRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_app_service_certificate_binding",
		RFunc: NewAppServiceCertificateBinding,
		ReferenceAttributes: []string{
			"certificate_id",
		},
	}
}
func NewAppServiceCertificateBinding(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.AppServiceCertificateBinding{Address: d.Address, Region: d.Get("region").String(), SSLState: lookupRegion(d, []string{"certificate_id"})}
	r.PopulateUsage(u)
	return r.BuildResource()
}
