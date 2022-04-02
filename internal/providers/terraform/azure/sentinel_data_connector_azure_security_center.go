package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func getSentinelDataConnectorAzureSecurityCenterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_sentinel_data_connector_azure_security_center",
		RFunc: newSentinelDataConnectorAzureSecurityCenter,
		ReferenceAttributes: []string{
			"log_analytics_workspace_id",
		},
	}
}

func newSentinelDataConnectorAzureSecurityCenter(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:        d.Address,
		IsSkipped:   true,
		NoPrice:     true,
		UsageSchema: []*schema.UsageItem{},
	}
}
