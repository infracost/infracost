package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMDNSmxRecordRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_dns_mx_record",
		RFunc: NewAzureRMDNSmxRecord,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func NewAzureRMDNSmxRecord(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: dnsQueriesCostComponent(d, u),
	}
}
