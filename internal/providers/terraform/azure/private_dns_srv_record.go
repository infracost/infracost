package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMPrivateDNSsrvRecordRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_private_dns_srv_record",
		RFunc: NewAzureRMPrivateDNSsrvRecord,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func NewAzureRMPrivateDNSsrvRecord(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: dnsQueriesCostComponent(d, u),
	}
}
