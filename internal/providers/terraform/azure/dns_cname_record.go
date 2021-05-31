package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMDNScnameRecordRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_dns_cname_record",
		RFunc: NewAzureRMDNScnameRecord,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func NewAzureRMDNScnameRecord(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	group := d.References("resource_group_name")[0]

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: DNSqueriesCostComponent(d, u, group),
	}
}
