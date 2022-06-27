package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getDNSNSRecordRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_dns_ns_record",
		RFunc: NewDNSNSRecord,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}
func NewDNSNSRecord(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.DNSNSRecord{Address: d.Address, Region: lookupRegion(d, []string{"resource_group_name"})}
	r.PopulateUsage(u)
	return r.BuildResource()
}
