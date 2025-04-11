package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAppServiceCertificateBindingRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_app_service_certificate_binding",
		RFunc: NewAppServiceCertificateBinding,
		ReferenceAttributes: []string{
			"certificate_id",
		},
		GetRegion: func(defaultRegion string, d *schema.ResourceData) string {
			return lookupRegion(d, []string{"certificate_id"})
		},
	}
}
func NewAppServiceCertificateBinding(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.AppServiceCertificateBinding{Address: d.Address, Region: d.Region, SSLState: d.Get("sslState").String()}
	r.PopulateUsage(u)
	return r.BuildResource()
}
