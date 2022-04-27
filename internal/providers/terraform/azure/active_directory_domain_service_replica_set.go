package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getActiveDirectoryDomainServiceReplicaSetRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_active_directory_domain_service_replica_set",
		RFunc: NewActiveDirectoryDomainServiceReplicaSet,
		ReferenceAttributes: []string{
			"domain_service_id",
		},
	}
}
func NewActiveDirectoryDomainServiceReplicaSet(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.ActiveDirectoryDomainServiceReplicaSet{
		Address: d.Address,
		Region:  lookupRegion(d, []string{}),
	}
	if len(d.References("domain_service_id")) > 0 {
		r.DomainServiceIDSKU = d.References("domain_service_id")[0].Get("sku").String()
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
