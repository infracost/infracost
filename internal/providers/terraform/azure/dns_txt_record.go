package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getDNSTxtRecordRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_dns_txt_record",
		RFunc: NewDNSTxtRecord,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}
func NewDNSTxtRecord(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.DNSTxtRecord{Address: d.Address, Region: lookupRegion(d, []string{"resource_group_name"})}
	r.PopulateUsage(u)
	return r.BuildResource()
}
