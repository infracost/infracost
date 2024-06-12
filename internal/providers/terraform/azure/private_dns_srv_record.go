package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getPrivateDNSSRVRecordRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_private_dns_srv_record",
		CoreRFunc: NewPrivateDNSSRVRecord,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}
func NewPrivateDNSSRVRecord(d *schema.ResourceData) schema.CoreResource {
	r := &azure.PrivateDNSSRVRecord{Address: d.Address, Region: d.Region}
	return r
}
