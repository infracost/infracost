package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMActiveDirectoryDomainServiceReplicaSetRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_active_directory_domain_service_replica_set",
		RFunc: NewAzureRMActiveDirectoryDomainServiceReplicaSet,
		ReferenceAttributes: []string{
			"domain_service_id",
		},
	}
}

func NewAzureRMActiveDirectoryDomainServiceReplicaSet(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{})
	var domainService *schema.ResourceData
	if len(d.References("domain_service_id")) > 0 {
		domainService = d.References("domain_service_id")[0]
	}
	costComponents := activeDirectoryDomainServiceCostComponents("Active directory domain service replica set", region, domainService)

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
