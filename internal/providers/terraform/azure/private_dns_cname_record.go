package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMPrivateDNScnameRecordRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_private_dns_cname_record",
		RFunc: NewAzureRMPrivateDNScnameRecord,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func NewAzureRMPrivateDNScnameRecord(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: dnsQueriesCostComponent(d, u),
	}
}
