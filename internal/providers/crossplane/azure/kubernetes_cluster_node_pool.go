package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

// getKubernetesClusterNodePoolRegistryItem ...
func getKubernetesClusterNodePoolRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "containerservice.azure.upbound.io/KubernetesClusterNodePool",
		CoreRFunc: NewKubernetesClusterNodePool,
		ReferenceAttributes: []string{
			"kubernetesClusterId",
		},
		GetRegion: func(defaultRegion string, d *schema.ResourceData) string {
			return lookupRegion(d, []string{"kubernetesClusterId"})
		},
	}
}

// NewKubernetesClusterNodePool ...
func NewKubernetesClusterNodePool(d *schema.ResourceData) schema.CoreResource {
	nodeCount := int64(1)
	if d.Get("forProvider.nodeCount").Type != gjson.Null {
		nodeCount = d.Get("forProvider.nodeCount").Int()
	}

	// if the node count is not set explicitly, let's take the min_count.
	if d.Get("forProvider.minCount").Type != gjson.Null && nodeCount == 1 {
		nodeCount = d.Get("forProvider.minCount").Int()
	}

	os := "Linux"
	if d.Get("forProvider.osType").Type != gjson.Null {
		os = d.Get("forProvider.osType").String()
	}

	if d.Get("forProvider.osSku").Type != gjson.Null {
		if strings.HasPrefix(strings.ToLower(d.Get("forProvider.osSku").String()), "windows") {
			os = "Windows"
		}
	}

	r := &azure.KubernetesClusterNodePool{
		Address:      d.Address,
		Region:       d.Region,
		VMSize:       d.Get("forProvider.vmSize").String(),
		OS:           os,
		OSDiskType:   d.Get("forProvider.osDiskType").String(),
		OSDiskSizeGB: d.Get("forProvider.osDiskSizeGb").Int(),
		NodeCount:    nodeCount,
	}

	return r
}
