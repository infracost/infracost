package azure

import (
	"strings"

	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getKubernetesClusterNodePoolRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_kubernetes_cluster_node_pool",
		CoreRFunc: NewKubernetesClusterNodePool,
		ReferenceAttributes: []string{
			"kubernetes_cluster_id",
		},
		GetRegion: func(defaultRegion string, d *schema.ResourceData) string {
			return lookupRegion(d, []string{"kubernetes_cluster_id"})
		},
	}
}

func NewKubernetesClusterNodePool(d *schema.ResourceData) schema.CoreResource {
	nodeCount := int64(1)
	if d.Get("node_count").Type != gjson.Null {
		nodeCount = d.Get("node_count").Int()
	}

	// if the node count is not set explicitly let's take the min_count.
	if d.Get("min_count").Type != gjson.Null && nodeCount == 1 {
		nodeCount = d.Get("min_count").Int()
	}

	os := "Linux"
	if d.Get("os_type").Type != gjson.Null {
		os = d.Get("os_type").String()
	}

	if d.Get("os_sku").Type != gjson.Null {
		if strings.HasPrefix(strings.ToLower(d.Get("os_sku").String()), "windows") {
			os = "Windows"
		}
	}

	r := &azure.KubernetesClusterNodePool{
		Address:      d.Address,
		Region:       d.Region,
		VMSize:       d.Get("vm_size").String(),
		OS:           os,
		OSDiskType:   d.Get("os_disk_type").String(),
		OSDiskSizeGB: d.Get("os_disk_size_gb").Int(),
		NodeCount:    nodeCount,
		IsDevTest:    d.ProjectMetadata["isProduction"] == "false",
	}
	return r
}
