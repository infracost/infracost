package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAppServiceCertificateBindingRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_app_service_certificate_binding",
		RFunc: NewAppServiceCertificateBinding,
		ReferenceAttributes: []string{
			"certificate_id",
		},
	}
}
func NewAppServiceCertificateBinding(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.AppServiceCertificateBinding{Address: d.Address, Region: lookupRegion(d, []string{"certificate_id"}), SSLState: d.Get("ssl_state").String()}
	r.PopulateUsage(u)
	return r.BuildResource()
}
