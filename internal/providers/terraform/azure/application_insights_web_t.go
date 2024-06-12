package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getApplicationInsightsWebTestRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_application_insights_web_test",
		CoreRFunc: NewApplicationInsightsWebTest,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}
func NewApplicationInsightsWebTest(d *schema.ResourceData) schema.CoreResource {
	r := &azure.ApplicationInsightsWebTest{
		Address: d.Address,
		Region:  d.Region,
		Enabled: d.Get("enabled").Bool(),
		Kind:    d.Get("kind").String(),
	}
	return r
}
