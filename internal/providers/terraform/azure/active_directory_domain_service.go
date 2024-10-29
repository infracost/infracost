package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getActiveDirectoryDomainServiceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_active_directory_domain_service",
		CoreRFunc: NewActiveDirectoryDomainService,
	}
}
func NewActiveDirectoryDomainService(d *schema.ResourceData) schema.CoreResource {
	r := &azure.ActiveDirectoryDomainService{Address: d.Address, Region: d.Region, SKU: d.Get("sku").String()}
	return r
}
