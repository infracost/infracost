package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getPrivateDNSMXRecordRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_private_dns_mx_record",
		CoreRFunc: NewPrivateDNSMXRecord,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}
func NewPrivateDNSMXRecord(d *schema.ResourceData) schema.CoreResource {
	r := &azure.PrivateDNSMXRecord{Address: d.Address, Region: d.Region}
	return r
}
