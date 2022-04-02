package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func getSentinelDataConnectorAzureActiveDirectoryRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_sentinel_data_connector_azure_active_directory",
		RFunc: newSentinelDataConnectorAzureActiveDirectory,
		ReferenceAttributes: []string{
			"log_analytics_workspace_id",
		},
	}
}

func newSentinelDataConnectorAzureActiveDirectory(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:        d.Address,
		IsSkipped:   true,
		NoPrice:     true,
		UsageSchema: []*schema.UsageItem{},
	}
}
