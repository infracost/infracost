package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMPrivateDNSptrRecordRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_private_dns_ptr_record",
		RFunc: NewAzureRMPrivateDNSptrRecord,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func NewAzureRMPrivateDNSptrRecord(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: dnsQueriesCostComponent(d, u),
	}
}
