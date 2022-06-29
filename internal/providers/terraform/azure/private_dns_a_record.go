package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getPrivateDNSARecordRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_private_dns_a_record",
		RFunc: NewPrivateDNSARecord,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}
func NewPrivateDNSARecord(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.PrivateDNSARecord{Address: d.Address, Region: lookupRegion(d, []string{"resource_group_name"})}
	r.PopulateUsage(u)
	return r.BuildResource()
}
