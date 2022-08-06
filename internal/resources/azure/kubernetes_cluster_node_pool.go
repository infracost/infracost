package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"strings"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type KubernetesClusterNodePool struct {
	Address      string
	Region       string
	NodeCount    int64
	VMSize       string
	OSDiskType   string
	OSDiskSizeGB int64
	Nodes        *int64 `infracost_usage:"nodes"`
}

var KubernetesClusterNodePoolUsageSchema = []*schema.UsageItem{{Key: "nodes", ValueType: schema.Int64, DefaultValue: 0}}

func (r *KubernetesClusterNodePool) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *KubernetesClusterNodePool) BuildResource() *schema.Resource {

	nodeCount := decimal.NewFromInt(1)
	if r.NodeCount != 0 {
		nodeCount = decimal.NewFromInt(r.NodeCount)
	}
	if r.Nodes != nil {
		nodeCount = decimal.NewFromInt(*r.Nodes)
	}

	return aksClusterNodePool(r.Address, r.Region, r.VMSize, r.OSDiskType, r.OSDiskSizeGB, nodeCount)
}

func aksClusterNodePool(name, region, instanceType, osDiskType string, osDiskSizeGB int64, nodeCount decimal.Decimal) *schema.Resource {
	var costComponents []*schema.CostComponent
	var subResources []*schema.Resource

	mainResource := &schema.Resource{
		Name: name,
	}
	costComponents = append(costComponents, linuxVirtualMachineCostComponent(region, instanceType, nil))
	mainResource.CostComponents = costComponents
	schema.MultiplyQuantities(mainResource, nodeCount)

	if osDiskType == "" {
		osDiskType = "Managed"
	}
	if strings.ToLower(osDiskType) == "managed" {
		diskSize := 128
		if osDiskSizeGB > 0 {
			diskSize = int(osDiskSizeGB)
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
