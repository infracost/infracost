package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getMonitorScheduledQueryRulesAlertRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_monitor_scheduled_query_rules_alert",
		CoreRFunc: newMonitorScheduledQueryRulesAlert,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func newMonitorScheduledQueryRulesAlert(d *schema.ResourceData) schema.CoreResource {
	region := d.Region
	return &azure.MonitorScheduledQueryRulesAlert{
		Address:          d.Address,
		Region:           region,
		Enabled:          d.GetBoolOrDefault("enabled", true),
		TimeSeriesCount:  int64(1),
		FrequencyMinutes: d.Get("frequency").Int(),
	}
}
