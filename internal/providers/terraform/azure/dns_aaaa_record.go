package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getDNSAAAARecordRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_dns_aaaa_record",
		RFunc: NewDNSAAAARecord,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}
func NewDNSAAAARecord(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.DNSAAAARecord{Address: d.Address, Region: lookupRegion(d, []string{"resource_group_name"})}
	r.PopulateUsage(u)
	return r.BuildResource()
}
