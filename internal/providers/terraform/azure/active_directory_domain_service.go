package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getActiveDirectoryDomainServiceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_active_directory_domain_service",
		RFunc: NewActiveDirectoryDomainService,
	}
}
func NewActiveDirectoryDomainService(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.ActiveDirectoryDomainService{Address: d.Address, Region: lookupRegion(d, []string{}), SKU: d.Get("sku").String()}
	r.PopulateUsage(u)
	return r.BuildResource()
}
