package azure

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetAzureRMKubernetesClusterNodePoolRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_kubernetes_cluster_node_pool",
		RFunc: NewAzureRMKubernetesClusterNodePool,
		ReferenceAttributes: []string{
			"kubernetes_cluster_id",
		},
	}
}

func NewAzureRMKubernetesClusterNodePool(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var costComponents []*schema.CostComponent
	var subResources []*schema.Resource

	mainResource := &schema.Resource{
		Name: d.Address,
	}

	mainCluster := d.References("kubernetes_cluster_id")
	var location string
	if len(mainCluster) > 0 {
		location = mainCluster[0].Get("location").String()
	}

	instanceType := d.Get("vm_size").String()
	nodeCount := decimal.NewFromInt(1)
	if d.Get("node_count").Type != gjson.Null {
		nodeCount = decimal.NewFromInt(d.Get("node_count").Int())
	}
	if u != nil && u.Get("nodes").Exists() {
		nodeCount = decimal.NewFromInt(u.Get("nodes").Int())
	}
	costComponents = append(costComponents, linuxVirtualMachineCostComponent(location, instanceType))
	mainResource.CostComponents = costComponents
	schema.MultiplyQuantities(mainResource, nodeCount)

	osDiskType := "Managed"
	if d.Get("os_disk_type").Type != gjson.Null {
		osDiskType = d.Get("os_disk_type").String()
	}
	if osDiskType == "Managed" {
		var diskSize int
		if d.Get("os_disk_size_gb").Type != gjson.Null {
			diskSize = int(d.Get("os_disk_size_gb").Int())
		}
		osDisk := aksOSDiskSubResource(location, diskSize, u)

		if osDisk != nil {
			subResources = append(subResources, osDisk)
			schema.MultiplyQuantities(subResources[0], nodeCount)
			mainResource.SubResources = subResources
		}
	}

	return mainResource
}
