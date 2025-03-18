package azure

import (
	"strings"

	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getKubernetesClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_kubernetes_cluster",
		CoreRFunc: NewKubernetesCluster,
	}
}

func NewKubernetesCluster(d *schema.ResourceData) schema.CoreResource {
	nodeCount := int64(1)
	if d.Get("default_node_pool.0.node_count").Type != gjson.Null {
		nodeCount = d.Get("default_node_pool.0.node_count").Int()
	}

	// if the node count is not set explicitly let's take the min_count.
	if d.Get("default_node_pool.0.min_count").Type != gjson.Null && nodeCount == 1 {
		nodeCount = d.Get("default_node_pool.0.min_count").Int()
	}

	os := "Linux"
	if d.Get("default_node_pool.0.os_type").Type != gjson.Null {
		os = d.Get("default_node_pool.0.os_type").String()
	}

	if d.Get("default_node_pool.0.os_sku").Type != gjson.Null {
		if strings.HasPrefix(strings.ToLower(d.Get("default_node_pool.0.os_sku").String()), "windows") {
			os = "Windows"
		}
	}

	r := &azure.KubernetesCluster{
		Address:                       d.Address,
		Region:                        d.Region,
		SKUTier:                       d.Get("sku_tier").String(),
		NetworkProfileLoadBalancerSKU: d.Get("network_profile.0.load_balancer_sku").String(),
		DefaultNodePoolNodeCount:      nodeCount,
		DefaultNodePoolOS:             os,
		DefaultNodePoolOSDiskType:     d.Get("default_node_pool.0.os_disk_type").String(),
		DefaultNodePoolVMSize:         d.Get("default_node_pool.0.vm_size").String(),
		DefaultNodePoolOSDiskSizeGB:   d.Get("default_node_pool.0.os_disk_size_gb").Int(),
		HttpApplicationRoutingEnabled: d.Get("http_application_routing_enabled").Bool(),
		IsDevTest:                     d.ProjectMetadata["isProduction"] == "false",
	}

	// Deprecated and removed in v3
	if !d.IsEmpty("addon_profile.0.http_application_routing") {
		r.HttpApplicationRoutingEnabled = d.Get("addon_profile.0.http_application_routing.0.enabled").Bool()
	}

	return r
}
