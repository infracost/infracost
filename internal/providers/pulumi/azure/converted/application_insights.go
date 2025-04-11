package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getApplicationInsightsRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_application_insights",
		RFunc: NewApplicationInsights,
	}
}
func NewApplicationInsights(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.ApplicationInsights{
		Address:         d.Address,
		Region:          d.Region,
		RetentionInDays: d.Get("retentionInDays").Int(),
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
