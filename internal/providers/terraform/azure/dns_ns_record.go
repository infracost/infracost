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
	group := d.References("resource_group_name")[0]

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: DNSqueriesCostComponent(d, u, group),
	}
}
