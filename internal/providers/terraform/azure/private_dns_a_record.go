package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMPrivateDNSaRecordRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_private_dns_a_record",
		RFunc: NewAzureRMPrivateDNSaRecord,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func NewAzureRMPrivateDNSaRecord(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: dnsQueriesCostComponent(d, u),
	}
}
