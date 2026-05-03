package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getContainerAppEnvironmentRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_container_app_environment",
		CoreRFunc: newContainerAppEnvironment,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func newContainerAppEnvironment(d *schema.ResourceData) schema.CoreResource {
	region := d.Region
	var workloadProfiles []azure.ContainerAppEnvironmentWorkloadProfile

	if d.Get("workload_profile").Exists() {
		for _, profile := range d.Get("workload_profile").Array() {
			workloadProfiles = append(workloadProfiles, azure.ContainerAppEnvironmentWorkloadProfile{
				Name:                profile.Get("name").String(),
				WorkloadProfileType: profile.Get("workload_profile_type").String(),
				MinimumCount:        profile.Get("minimum_count").Int(),
				MaximumCount:        profile.Get("maximum_count").Int(),
			})
		}
	}

	return &azure.ContainerAppEnvironment{
		Address:          d.Address,
		Region:           region,
		WorkloadProfiles: workloadProfiles,
	}
}
