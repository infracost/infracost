package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getApplicationGatewayRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_application_gateway",
		RFunc: NewApplicationGateway,
	}
}
func NewApplicationGateway(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.ApplicationGateway{Address: d.Address, Region: lookupRegion(d, []string{}), SKUName: d.Get("sku.0.name").String(), SKUCapacity: d.Get("sku.0.capacity").Int()}
	r.PopulateUsage(u)
	return r.BuildResource()
}
