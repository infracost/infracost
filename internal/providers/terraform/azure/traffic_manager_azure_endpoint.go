package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getTrafficManagerAzureEndpointRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_traffic_manager_azure_endpoint",
		CoreRFunc: newTrafficManagerAzureEndpoint,
		ReferenceAttributes: []string{
			"profile_id",
		},
	}
}

func newTrafficManagerAzureEndpoint(d *schema.ResourceData) schema.CoreResource {
	region := ""
	healthCheckInterval := int64(30)
	profileEnabled := false

	if len(d.References("profile_id")) > 0 {
		profile := d.References("profile_id")[0]
		region = lookupRegion(profile, []string{"resource_group_name"})
		healthCheckInterval = profile.GetInt64OrDefault("monitor_config.0.interval_in_seconds", 30)
		profileEnabled = trafficManagerProfileEnabled(profile)
	}

	return &azure.TrafficManagerEndpoint{
		Address:             d.Address,
		Region:              region,
		ProfileEnabled:      profileEnabled,
		External:            false,
		HealthCheckInterval: healthCheckInterval,
	}
}
