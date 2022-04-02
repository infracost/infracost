package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func getSentinelSentinelDataConnectorThreatIntelligenceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_sentinel_sentinel_data_connector_threat_intelligence",
		RFunc: newSentinelSentinelDataConnectorThreatIntelligence,
		ReferenceAttributes: []string{
			"log_analytics_workspace_id",
		},
	}
}

func newSentinelSentinelDataConnectorThreatIntelligence(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:        d.Address,
		IsSkipped:   true,
		NoPrice:     true,
		UsageSchema: []*schema.UsageItem{},
	}
}
