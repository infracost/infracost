package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getMachineLearningComputeClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_machine_learning_compute_cluster",
		CoreRFunc: newMachineLearningComputeCluster,
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
		InstanceType: d.Get("vm_size").String(),
		MinNodeCount: d.Get("scale_settings.0.min_node_count").Int(),
	}
}
