package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMDNSnsRecordRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_dns_ns_record",
		RFunc: NewAzureRMDNSnsRecord,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func NewAzureRMDNSnsRecord(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: dnsQueriesCostComponent(d, u),
	}
}
