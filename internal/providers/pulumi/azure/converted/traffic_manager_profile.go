package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
	"strings"
)

func getTrafficManagerProfileRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_traffic_manager_profile",
		RFunc: newTrafficManagerProfile,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func newTrafficManagerProfile(d *schema.ResourceData) schema.CoreResource {
	region := d.Region

	return &azure.TrafficManagerProfile{
		Address:            d.Address,
		Region:             region,
		Enabled:            trafficManagerProfileEnabled(d),
		TrafficViewEnabled: d.Get("trafficViewEnabled").Bool(),
	}
}

func trafficManagerProfileEnabled(d *schema.ResourceData) bool {
	return strings.ToLower(d.GetStringOrDefault("profileStatus", "enabled")) == "enabled"
}
