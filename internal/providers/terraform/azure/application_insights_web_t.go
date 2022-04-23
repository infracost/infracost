package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getApplicationInsightsWebTestRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_application_insights_web_test",
		RFunc: NewApplicationInsightsWebTest,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}
func NewApplicationInsightsWebTest(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.ApplicationInsightsWebTest{
		Address: d.Address,
		Region:  lookupRegion(d, []string{"resource_group_name"}),
		Enabled: d.Get("enabled").Bool(),
		Kind:    d.Get("kind").String(),
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
