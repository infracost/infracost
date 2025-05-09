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
		RFunc: NewKubernetesClusterNodePool,
		ReferenceAttributes: []string{
			"kubernetes_cluster_id",
		},
		GetRegion: func(defaultRegion string, d *schema.ResourceData) string {
			return lookupRegion(d, []string{"kubernetes_cluster_id"})
		},
	}
}

func NewKubernetesClusterNodePool(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	nodeCount := int64(1)
	if d.Get("nodeCount").Type != gjson.Null {
		nodeCount = d.Get("nodeCount").Int()
	}

	// if the node count is not set explicitly let's take the min_count.
	if d.Get("minCount").Type != gjson.Null && nodeCount == 1 {
		nodeCount = d.Get("minCount").Int()
	}

	os := "Linux"
	if d.Get("osType").Type != gjson.Null {
		os = d.Get("osType").String()
	}

	if d.Get("osSku").Type != gjson.Null {
		if strings.HasPrefix(strings.ToLower(d.Get("osSku").String()), "windows") {
			os = "Windows"
		}
	}

	r := &azure.KubernetesClusterNodePool{
		Address:      d.Address,
		Region:       d.Region,
		VMSize:       d.Get("vmSize").String(),
		OS:           os,
		OSDiskType:   d.Get("osDiskType").String(),
		OSDiskSizeGB: d.Get("osDiskSizeGb").Int(),
		NodeCount:    nodeCount,
		IsDevTest:    d.ProjectMetadata["isProduction"] == "false",
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
