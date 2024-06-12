package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getNetworkWatcherFlowLogRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_network_watcher_flow_log",
		CoreRFunc: newNetworkWatcherFlowLog,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func newNetworkWatcherFlowLog(d *schema.ResourceData) schema.CoreResource {
	if !d.Get("enabled").Bool() {
		return schema.BlankCoreResource{
			Name: d.Address,
			Type: d.Type,
		}
	}

	trafficAnalyticsEnabled := false
	trafficAnalyticsAcceleratedProcessing := false

	if len(d.Get("traffic_analytics").Array()) > 0 {
		trafficAnalyticsEnabled = d.Get("traffic_analytics.0.enabled").Bool()
		trafficAnalyticsAcceleratedProcessing = d.Get("traffic_analytics.0.interval_in_minutes").Int() == int64(10)
	}

	region := d.Region
	return &azure.NetworkWatcherFlowLog{
		Address:                               d.Address,
		Region:                                region,
		TrafficAnalyticsEnabled:               trafficAnalyticsEnabled,
		TrafficAnalyticsAcceleratedProcessing: trafficAnalyticsAcceleratedProcessing,
	}
}
