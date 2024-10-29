package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getNetworkConnectionMonitorRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_network_connection_monitor",
		CoreRFunc: newNetworkConnectionMonitor,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func newNetworkConnectionMonitor(d *schema.ResourceData) schema.CoreResource {
	tests := 0

	for _, testGroup := range d.Get("test_group").Array() {
		if !testGroup.Get("enabled").Exists() || testGroup.Get("enabled").Bool() {
			destinationCount := len(testGroup.Get("destination_endpoints").Array())
			sourceCount := len(testGroup.Get("source_endpoints").Array())
			testConfigCount := len(testGroup.Get("test_configuration_names").Array())
			tests += sourceCount * destinationCount * testConfigCount
		}
	}

	region := d.Region
	return &azure.NetworkConnectionMonitor{
		Address: d.Address,
		Region:  region,
		Tests:   intPtr(int64(tests)),
	}
}
