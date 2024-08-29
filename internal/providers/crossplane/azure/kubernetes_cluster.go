package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

// getKubernetesClusterRegistryItem ...
func getKubernetesClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "containerservice.azure.upbound.io/KubernetesCluster",
		CoreRFunc: NewKubernetesCluster,
	}
}

// Reference: https://marketplace.upbound.io/providers/upbound/provider-azure-containerservice/v1.5.0
func NewKubernetesCluster(d *schema.ResourceData) schema.CoreResource {
	forProvider := d.Get("forProvider")
	logging.Logger.Debug().Msgf("Parsing forProvider: %s", forProvider.Raw)

	nodeCount := int64(1)
	if forProvider.Get("defaultNodePool.0.nodeCount").Type != gjson.Null {
		nodeCount = forProvider.Get("defaultNodePool.0.nodeCount").Int()
		logging.Logger.Debug().Msgf("Detected Node Count: %d", nodeCount)
	}

	// if the node count is not set explicitly let's take the min_count.
	if forProvider.Get("defaultNodePool.0.minCount").Type != gjson.Null && nodeCount == 1 {
		nodeCount = forProvider.Get("defaultNodePool.0.minCount").Int()
		logging.Logger.Debug().Msgf("Detected Min Node Count: %d", nodeCount)
	}

	os := "Linux"
	if forProvider.Get("defaultNodePool.0.osSku").Type != gjson.Null {
		if strings.HasPrefix(strings.ToLower(forProvider.Get("defaultNodePool.0.osSku").String()), "windows") {
			os = "Windows"
		}
		logging.Logger.Debug().Msgf("Detected OS: %s", os)
	}

	skuTier := "Free"
	if forProvider.Get("skuTier").Type != gjson.Null {
		skuTier = strings.Title(strings.ToLower(forProvider.Get("skuTier").String()))
		logging.Logger.Debug().Msgf("Detected SKU Tier: %s", skuTier)
	}

	region := lookupRegion(d, []string{})

	r := &azure.KubernetesCluster{
		Address:                       d.Address,
		Region:                        region,
		SKUTier:                       skuTier,
		NetworkProfileLoadBalancerSKU: forProvider.Get("networkProfile.0.loadBalancerSku").String(),
		DefaultNodePoolNodeCount:      nodeCount,
		DefaultNodePoolOS:             os,
		DefaultNodePoolOSDiskType:     forProvider.Get("defaultNodePool.0.osDiskType").String(),
		DefaultNodePoolVMSize:         forProvider.Get("defaultNodePool.0.vmSize").String(),
		DefaultNodePoolOSDiskSizeGB:   forProvider.Get("defaultNodePool.0.osDiskSizeGb").Int(),
		HttpApplicationRoutingEnabled: forProvider.Get("httpApplicationRoutingEnabled").Bool(),
	}

	logging.Logger.Debug().Msgf("Created KubernetesCluster: %s", fmt.Sprintf("%v", r))

	return r
}
