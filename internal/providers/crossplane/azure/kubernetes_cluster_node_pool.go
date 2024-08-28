package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

// GetAzureRMKubernetesClusterNodePoolRegistryItem ...
func GetAzureRMKubernetesClusterNodePoolRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "containerservice.azure.upbound.io/v1beta1",
		RFunc: NewAzureRMKubernetesClusterNodePool,
		ReferenceAttributes: []string{
			"kubernetesClusterId",
		},
	}
}

// NewAzureRMKubernetesClusterNodePool ...
func NewAzureRMKubernetesClusterNodePool(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{"kubernetesClusterId"})

	nodeCount := decimal.NewFromInt(1)
	if d.Get("forProvider.nodeCount").Type != gjson.Null {
		nodeCount = decimal.NewFromInt(d.Get("forProvider.nodeCount").Int())
	}

	return aksClusterNodePool(d.Address, region, d.RawValues, nodeCount, u)
}

func aksClusterNodePool(name, region string, n gjson.Result, nodeCount decimal.Decimal, u *schema.UsageData) *schema.Resource {
	var costComponents []*schema.CostComponent
	var subResources []*schema.Resource

	mainResource := &schema.Resource{
		Name: name,
	}
	instanceType := n.Get("forProvider.vmSize").String()
	costComponents = append(costComponents, linuxVirtualMachineCostComponent(region, instanceType))
	mainResource.CostComponents = costComponents
	schema.MultiplyQuantities(mainResource, nodeCount)

	osDiskType := "Managed"
	if n.Get("forProvider.osDiskType").Type != gjson.Null {
		osDiskType = n.Get("forProvider.osDiskType").String()
	}
	if strings.ToLower(osDiskType) == "managed" {
		var diskSize int
		if n.Get("forProvider.osDiskSizeGb").Type != gjson.Null {
			diskSize = int(n.Get("forProvider.osDiskSizeGb").Int())
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
