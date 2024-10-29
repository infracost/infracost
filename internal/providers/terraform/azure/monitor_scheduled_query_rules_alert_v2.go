package azure

import (
	duration "github.com/channelmeter/iso8601duration"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getMonitorScheduledQueryRulesAlertV2RegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_monitor_scheduled_query_rules_alert_v2",
		CoreRFunc: newMonitorScheduledQueryRulesAlertV2,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func newMonitorScheduledQueryRulesAlertV2(d *schema.ResourceData) schema.CoreResource {
	region := d.Region

	freq := int64(1)
	ef, err := duration.FromString(d.Get("evaluation_frequency").String())
	if err != nil {
		logging.Logger.Warn().Str(
			"resource", d.Address,
		).Msgf("failed to parse ISO8061 duration string '%s' using 1 minute frequency", d.Get("evaluation_frequency").String())
	} else {
		freq = int64(ef.ToDuration().Minutes())
	}

	scopeCount := 1 // default scope is the azure subscription, so count == 1
	if !d.IsEmpty("scopes") {
		scopeCount = len(d.Get("scopes").Array())
	}

	criteriaDimensionsCount := 0
	for _, c := range d.Get("criteria").Array() {
		criteriaDimensionsCount += len(c.Get("dimension").Array())
	}

	return &azure.MonitorScheduledQueryRulesAlert{
		Address:          d.Address,
		Region:           region,
		Enabled:          d.GetBoolOrDefault("enabled", true),
		TimeSeriesCount:  int64(scopeCount * criteriaDimensionsCount),
		FrequencyMinutes: freq,
	}
}
