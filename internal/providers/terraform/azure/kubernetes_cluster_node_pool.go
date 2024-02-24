package azure

import (
	"regexp"
	"strings"

	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

func getKubernetesClusterNodePoolRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_kubernetes_cluster_node_pool",
		RFunc: NewKubernetesClusterNodePool,
		ReferenceAttributes: []string{
			"kubernetes_cluster_id",
		},
	}
}

func NewKubernetesClusterNodePool(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
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
		Region:       lookupRegion(d, []string{"kubernetes_cluster_id"}),
		VMSize:       d.Get("vm_size").String(),
		OS:           os,
		OSDiskType:   d.Get("os_disk_type").String(),
		OSDiskSizeGB: d.Get("os_disk_size_gb").Int(),
		NodeCount:    nodeCount,
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}

func aksGetStorageType(instanceType string) string {
	parts := strings.Split(instanceType, "_")

	subfamily := ""
	if len(parts) > 1 {
		subfamily = parts[1]
	}

	// Check if the subfamily is a known premium type
	premiumPrefixes := []string{"ds", "gs", "m"}
	for _, p := range premiumPrefixes {
		if strings.HasPrefix(strings.ToLower(subfamily), p) {
			return "Premium"
		}
	}

	// Otherwise check if it contains an s as an 'Additive Feature'
	// as per https://learn.microsoft.com/en-us/azure/virtual-machines/vm-naming-conventions
	re := regexp.MustCompile(`\d+[A-Za-z]*(s)`)
	matches := re.FindStringSubmatch(subfamily)

	if len(matches) > 0 {
		return "Premium"
	}

	return "Standard"
}
