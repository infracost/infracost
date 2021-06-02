package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMDNSptrRecordRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_dns_ptr_record",
		RFunc: NewAzureRMDNSptrRecord,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func NewAzureRMDNSptrRecord(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: dnsQueriesCostComponent(d, u),
	}
}
