package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getDNSAAAARecordRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_dns_aaaa_record",
		CoreRFunc: NewDNSAAAARecord,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}
func NewDNSAAAARecord(d *schema.ResourceData) schema.CoreResource {
	r := &azure.DNSAAAARecord{Address: d.Address, Region: d.Region}
	return r
}
