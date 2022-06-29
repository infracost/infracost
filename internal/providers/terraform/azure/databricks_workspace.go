package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getDatabricksWorkspaceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_databricks_workspace",
		RFunc: NewDatabricksWorkspace,
	}
}
func NewDatabricksWorkspace(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.DatabricksWorkspace{Address: d.Address, Region: lookupRegion(d, []string{}), SKU: d.Get("sku").String()}
	r.PopulateUsage(u)
	return r.BuildResource()
}
