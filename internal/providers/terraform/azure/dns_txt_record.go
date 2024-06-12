package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getDNSTxtRecordRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_dns_txt_record",
		CoreRFunc: NewDNSTxtRecord,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}
func NewDNSTxtRecord(d *schema.ResourceData) schema.CoreResource {
	r := &azure.DNSTxtRecord{Address: d.Address, Region: d.Region}
	return r
}
