package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getAzureKubernetesClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_kubernetes_cluster",
		RFunc: NewKubernetesCluster,
	}
}
func NewKubernetesCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.KubernetesCluster{
		Address:                       d.Address,
		Region:                        lookupRegion(d, []string{}),
		SKUTier:                       d.Get("sku_tier").String(),
		NetworkProfileLoadBalancerSKU: d.Get("network_profile.0.load_balancer_sku").String(),
		DefaultNodePoolNodeCount:      d.Get("default_node_pool.0.node_count").Int(),
		DefaultNodePoolOSDiskType:     d.Get("default_node_pool.0.os_disk_type").String(),
		DefaultNodePoolVMSize:         d.Get("default_node_pool.0.vm_size").String(),
		DefaultNodePoolOSDiskSizeGB:   d.Get("default_node_pool.0.os_disk_size_gb").Int(),
		HttpApplicationRoutingEnabled: d.Get("http_application_routing_enabled").Bool(),
	}

	// Deprecated and removed in v3
	if !d.IsEmpty("addon_profile.0.http_application_routing") {
		r.HttpApplicationRoutingEnabled = d.Get("addon_profile.0.http_application_routing.0.enabled").Bool()
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
