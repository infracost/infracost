package azure

import (
	"github.com/infracost/infracost/internal/schema"
)

func getSentinelDataConnectorThreatIntelligenceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_sentinel_data_connector_threat_intelligence",
		RFunc: newSentinelDataConnectorThreatIntelligence,
		ReferenceAttributes: []string{
			"log_analytics_workspace_id",
		},
	}
}

func newSentinelDataConnectorThreatIntelligence(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	return &schema.Resource{
		Name:        d.Address,
		IsSkipped:   true,
		NoPrice:     true,
		UsageSchema: []*schema.UsageItem{},
	}
}
