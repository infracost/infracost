package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getDNSZoneRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_dns_zone",
		CoreRFunc: NewDNSZone,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
		Notes: []string{"Most expensive price tier is used."},
	}
}

func NewDNSZone(d *schema.ResourceData) schema.CoreResource {
	r := &azure.DNSZone{
		Address: d.Address,
		Region:  d.Region,
	}

	return r
}
