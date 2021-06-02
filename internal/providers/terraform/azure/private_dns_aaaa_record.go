package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMPrivateDNSaaaaRecordRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_private_dns_aaaa_record",
		RFunc: NewAzureRMPrivateDNSaaaaRecord,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func NewAzureRMPrivateDNSaaaaRecord(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: dnsQueriesCostComponent(d, u),
	}
}
