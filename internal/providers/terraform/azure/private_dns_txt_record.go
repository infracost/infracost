package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getPrivateDNSTXTRecordRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_private_dns_txt_record",
		RFunc: NewPrivateDNSTXTRecord,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}
func NewPrivateDNSTXTRecord(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.PrivateDNSTXTRecord{Address: d.Address, Region: lookupRegion(d, []string{"resource_group_name"})}
	r.PopulateUsage(u)
	return r.BuildResource()
}
