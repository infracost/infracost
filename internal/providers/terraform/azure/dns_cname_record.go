package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getDNSCNameRecordRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_dns_cname_record",
		CoreRFunc: NewDNSCNameRecord,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}
func NewDNSCNameRecord(d *schema.ResourceData) schema.CoreResource {
	r := &azure.DNSCNameRecord{Address: d.Address, Region: d.Region}
	return r
}
