package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getMonitorMetricAlertRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_monitor_metric_alert",
		CoreRFunc: newMonitorMetricAlert,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func newMonitorMetricAlert(d *schema.ResourceData) schema.CoreResource {
	region := d.Region

	scopeCount := 1 // default scope is the azure subscription, so count == 1
	if !d.IsEmpty("scopes") {

		scopeCount = len(d.Get("scopes").Array())
	}

	criteriaDimensionsCount := 0
	for _, c := range d.Get("criteria").Array() {
		criteriaDimensionsCount += len(c.Get("dimension").Array())
	}

	dynamicCriteriaDimensionsCount := 0
	for _, c := range d.Get("dynamic_criteria").Array() {
		dynamicCriteriaDimensionsCount += len(c.Get("dimension").Array())
	}

	return &azure.MonitorMetricAlert{
		Address:                        d.Address,
		Region:                         region,
		Enabled:                        d.GetBoolOrDefault("enabled", true),
		ScopeCount:                     scopeCount,
		CriteriaDimensionsCount:        criteriaDimensionsCount,
		DynamicCriteriaDimensionsCount: dynamicCriteriaDimensionsCount,
	}
}
