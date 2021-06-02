package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMDNSaaaaRecordRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_dns_aaaa_record",
		RFunc: NewAzureRMDNSaaaaRecord,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func NewAzureRMDNSaaaaRecord(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: dnsQueriesCostComponent(d, u),
	}
}
