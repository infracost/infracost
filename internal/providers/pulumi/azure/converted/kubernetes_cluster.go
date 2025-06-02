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
		RFunc: NewKubernetesCluster,
	}
}

func NewKubernetesCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	nodeCount := int64(1)
	if d.Get("defaultNodePool.0.nodeCount").Type != gjson.Null {
		nodeCount = d.Get("defaultNodePool.0.nodeCount").Int()
	}

	// if the node count is not set explicitly let's take the min_count.
	if d.Get("defaultNodePool.0.minCount").Type != gjson.Null && nodeCount == 1 {
		nodeCount = d.Get("defaultNodePool.0.minCount").Int()
	}

	os := "Linux"
	if d.Get("defaultNodePool.0.osType").Type != gjson.Null {
		os = d.Get("defaultNodePool.0.osType").String()
	}

	if d.Get("defaultNodePool.0.osSku").Type != gjson.Null {
		if strings.HasPrefix(strings.ToLower(d.Get("defaultNodePool.0.osSku").String()), "windows") {
			os = "Windows"
		}
	}

	r := &azure.KubernetesCluster{
		Address:                       d.Address,
		Region:                        d.Region,
		SKUTier:                       d.Get("skuTier").String(),
		NetworkProfileLoadBalancerSKU: d.Get("networkProfile.0.loadBalancerSku").String(),
		DefaultNodePoolNodeCount:      nodeCount,
		DefaultNodePoolOS:             os,
		DefaultNodePoolOSDiskType:     d.Get("defaultNodePool.0.osDiskType").String(),
		DefaultNodePoolVMSize:         d.Get("defaultNodePool.0.vmSize").String(),
		DefaultNodePoolOSDiskSizeGB:   d.Get("defaultNodePool.0.osDiskSizeGb").Int(),
		HttpApplicationRoutingEnabled: d.Get("httpApplicationRoutingEnabled").Bool(),
		IsDevTest:                     d.ProjectMetadata["isProduction"] == "false",
	}

	// Deprecated and removed in v3
	if !d.IsEmpty("addon_profile.0.http_application_routing") {
		r.HttpApplicationRoutingEnabled = d.Get("addonProfile.0.httpApplicationRouting.0.enabled").Bool()
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
