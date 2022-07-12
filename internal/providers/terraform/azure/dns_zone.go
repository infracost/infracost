package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAzureDNSZoneRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_dns_zone",
		RFunc: NewDNSZone,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
		Notes: []string{"Most expensive price tier is used."},
	}
}
func NewDNSZone(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.DNSZone{Address: d.Address, Region: lookupRegion(d, []string{"resource_group_name"})}
	r.PopulateUsage(u)
	return r.BuildResource()
}
