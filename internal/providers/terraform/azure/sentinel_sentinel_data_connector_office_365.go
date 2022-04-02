package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func getSentinelSentinelDataConnectorOffice365RegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_sentinel_sentinel_data_connector_office_365",
		RFunc: newSentinelSentinelDataConnectorOffice365,
		ReferenceAttributes: []string{
			"log_analytics_workspace_id",
		},
	}
}

func newSentinelSentinelDataConnectorOffice365(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:        d.Address,
		IsSkipped:   true,
		NoPrice:     true,
		UsageSchema: []*schema.UsageItem{},
	}
}
