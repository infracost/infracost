package azure

import (
	"strings"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/schema"
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
	region := lookupRegion(d, []string{"kubernetes_cluster_id"})

	nodeCount := decimal.NewFromInt(1)
	if d.Get("node_count").Type != gjson.Null {
		nodeCount = decimal.NewFromInt(d.Get("node_count").Int())
	}

	// if the node count is not set explicitly let's take the min_count.
	if d.Get("min_count").Type != gjson.Null && nodeCount.Equal(decimal.NewFromInt(1)) {
		nodeCount = decimal.NewFromInt(d.Get("min_count").Int())
	}

	if u != nil && u.Get("nodes").Exists() {
		nodeCount = decimal.NewFromInt(u.Get("nodes").Int())
	}

	return aksClusterNodePool(d.Address, region, d.RawValues, nodeCount, u)
}

func aksClusterNodePool(name, region string, n gjson.Result, nodeCount decimal.Decimal, u *schema.UsageData) *schema.Resource {
	var costComponents []*schema.CostComponent
	var subResources []*schema.Resource

	mainResource := &schema.Resource{
		Name: name,
	}
	instanceType := n.Get("vm_size").String()
	costComponents = append(costComponents, linuxVirtualMachineCostComponent(region, instanceType, nil))
	mainResource.CostComponents = costComponents
	schema.MultiplyQuantities(mainResource, nodeCount)

	osDiskType := "Managed"
	if n.Get("os_disk_type").Type != gjson.Null {
		osDiskType = n.Get("os_disk_type").String()
	}
	if strings.ToLower(osDiskType) == "managed" {
		diskSize := 128
		if n.Get("os_disk_size_gb").Type != gjson.Null {
			diskSize = int(n.Get("os_disk_size_gb").Int())
		}
		osDisk := aksOSDiskSubResource(region, diskSize)

		if osDisk != nil {
			subResources = append(subResources, osDisk)
			schema.MultiplyQuantities(subResources[0], nodeCount)
			mainResource.SubResources = subResources
		}
	}

	return mainResource
}

func aksOSDiskSubResource(region string, diskSize int) *schema.Resource {
	diskType := "Premium_LRS"

	diskName := mapDiskName(diskType, diskSize)
	if diskName == "" {
		log.Warnf("Could not map disk type %s and size %d to disk name", diskType, diskSize)
		return nil
	}

	productName, ok := diskProductNameMap[diskType]
	if !ok {
		log.Warnf("Could not map disk type %s to product name", diskType)
		return nil
	}

	costComponent := []*schema.CostComponent{storageCostComponent(region, diskName, productName)}

	return &schema.Resource{
		Name:           "os_disk",
		CostComponents: costComponent,
	}
}
