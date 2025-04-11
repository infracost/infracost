package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func getSentinelDataConnectorMicrosoftCloudAppSecurityRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_sentinel_data_connector_microsoft_cloud_app_security",
		RFunc: newSentinelDataConnectorMicrosoftCloudAppSecurity,
		ReferenceAttributes: []string{
			"log_analytics_workspace_id",
		},
	}
}

func newSentinelDataConnectorMicrosoftCloudAppSecurity(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:        d.Address,
		IsSkipped:   true,
		NoPrice:     true,
		UsageSchema: []*schema.UsageItem{},
	}

}
