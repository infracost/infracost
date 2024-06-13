package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getTrafficManagerExternalEndpointRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_traffic_manager_external_endpoint",
		CoreRFunc: newTrafficManagerExternalEndpoint,
		ReferenceAttributes: []string{
			"profile_id",
		},
		GetRegion: func(defaultRegion string, d *schema.ResourceData) string {
			if len(d.References("profile_id")) > 0 {
				profile := d.References("profile_id")[0]
				return lookupRegion(profile, []string{"resource_group_name"})
			}

			return ""
		},
	}
}

func newTrafficManagerExternalEndpoint(d *schema.ResourceData) schema.CoreResource {
	region := d.Region
	healthCheckInterval := int64(30)
	profileEnabled := false

	if len(d.References("profile_id")) > 0 {
		profile := d.References("profile_id")[0]
		healthCheckInterval = profile.GetInt64OrDefault("monitor_config.0.interval_in_seconds", 30)
		profileEnabled = trafficManagerProfileEnabled(profile)
	}

	return &azure.TrafficManagerEndpoint{
		Address:             d.Address,
		Region:              region,
		ProfileEnabled:      profileEnabled,
		External:            true,
		HealthCheckInterval: healthCheckInterval,
	}
}
