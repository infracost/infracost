package azure

import (
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getContainerAppRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_container_app",
		CoreRFunc: newContainerApp,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func newContainerApp(d *schema.ResourceData) schema.CoreResource {
	region := d.Region

	var totalvCPU float64
	var totalMemory float64
	var minReplicas int64

	if d.Get("template").Exists() {
		for _, template := range d.Get("template").Array() {
			minReplicas = template.Get("min_replicas").Int()
			for _, container := range template.Get("container").Array() {
				totalvCPU += container.Get("cpu").Float()

				memStr := container.Get("memory").String()
				if memStr != "" {
					q, err := resource.ParseQuantity(memStr)
					if err == nil {
						// Convert bytes to GiB
						// Value() returns int64 bytes
						val := q.Value()
						totalMemory += float64(val) / 1024 / 1024 / 1024
					}
				}
			}
		}
	}

	return &azure.ContainerApp{
		Address:             d.Address,
		Region:              region,
		WorkloadProfileName: d.Get("workload_profile_name").String(),
		TotalvCPU:           totalvCPU,
		TotalMemory:         totalMemory,
		MinReplicas:         minReplicas,
	}
}
