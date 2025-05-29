package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getMachineLearningComputeClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_machine_learning_compute_cluster",
		RFunc: newMachineLearningComputeCluster,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func newMachineLearningComputeCluster(d *schema.ResourceData) schema.CoreResource {
	region := d.Region
	return &azure.MachineLearningComputeCluster{
		Address:      d.Address,
		Region:       region,
		InstanceType: d.Get("vmSize").String(),
		MinNodeCount: d.Get("scaleSettings.0.minNodeCount").Int(),
	}
}
