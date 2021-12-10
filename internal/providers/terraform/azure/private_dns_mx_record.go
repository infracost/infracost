package azure

import (
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMPrivateDNSmxRecordRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_private_dns_mx_record",
		RFunc: NewAzureRMPrivateDNSmxRecord,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func NewAzureRMPrivateDNSmxRecord(ctx *config.ProjectContext, d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:           d.Address,
		CostComponents: dnsQueriesCostComponent(d, u),
	}
}
