package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getPrivateDNSTXTRecordRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_private_dns_txt_record",
		CoreRFunc: NewPrivateDNSTXTRecord,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}
func NewPrivateDNSTXTRecord(d *schema.ResourceData) schema.CoreResource {
	r := &azure.PrivateDNSTXTRecord{Address: d.Address, Region: d.Region}
	return r
}
