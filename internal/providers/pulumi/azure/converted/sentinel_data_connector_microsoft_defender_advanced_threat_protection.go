package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func getSentinelDataConnectorMicrosoftDefenderAdvancedThreatProtectionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_sentinel_data_connector_microsoft_defender_advanced_threat_protection",
		RFunc: newSentinelDataConnectorMicrosoftDefenderAdvancedThreatProtection,
		ReferenceAttributes: []string{
			"log_analytics_workspace_id",
		},
	}
}

func newSentinelDataConnectorMicrosoftDefenderAdvancedThreatProtection(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:        d.Address,
		IsSkipped:   true,
		NoPrice:     true,
		UsageSchema: []*schema.UsageItem{},
	}
}
