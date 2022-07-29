package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAzureKubernetesClusterNodePoolRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_kubernetes_cluster_node_pool",
		RFunc: NewKubernetesClusterNodePool,
		ReferenceAttributes: []string{
			"kubernetes_cluster_id",
		},
	}
}
func NewKubernetesClusterNodePool(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.KubernetesClusterNodePool{
		Address:      d.Address,
		Region:       lookupRegion(d, []string{"kubernetes_cluster_id"}),
		VMSize:       d.Get("vm_size").String(),
		OSDiskType:   d.Get("os_disk_type").String(),
		OSDiskSizeGB: d.Get("os_disk_size_gb").Int(),
		NodeCount:    d.Get("node_count").Int(),
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
