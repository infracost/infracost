package azure

import (
	"regexp"
	"strings"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type KubernetesClusterNodePool struct {
	Address      string
	Region       string
	NodeCount    int64
	VMSize       string
	OS           string
	OSDiskType   string
	OSDiskSizeGB int64
	Nodes        *int64   `infracost_usage:"nodes"`
	MonthlyHours *float64 `infracost_usage:"monthly_hrs"`
	IsDevTest    bool
}

func (r *KubernetesClusterNodePool) CoreType() string {
	return "KubernetesClusterNodePool"
}

func (r *KubernetesClusterNodePool) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "nodes", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_hrs", ValueType: schema.Float64, DefaultValue: 0},
	}
}

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

	pool := aksClusterNodePool(r.Address, r.Region, r.VMSize, r.OS, r.OSDiskType, r.OSDiskSizeGB, nodeCount, r.MonthlyHours, r.IsDevTest)
	pool.UsageSchema = r.UsageSchema()
	return pool
}

func aksClusterNodePool(name, region, instanceType, os string, osDiskType string, osDiskSizeGB int64, nodeCount decimal.Decimal, monthlyHours *float64, isDevTest bool) *schema.Resource {
	var costComponents []*schema.CostComponent
	var subResources []*schema.Resource

	mainResource := &schema.Resource{
		Name: name,
	}

	if strings.EqualFold(os, "windows") {
		costComponents = append(costComponents, windowsVirtualMachineCostComponent(region, instanceType, "None", monthlyHours, isDevTest))
	} else {
		costComponents = append(costComponents, linuxVirtualMachineCostComponent(region, instanceType, monthlyHours))
	}

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
		osDisk := aksOSDiskSubResource(region, diskSize, instanceType)

		if osDisk != nil {
			subResources = append(subResources, osDisk)
			schema.MultiplyQuantities(subResources[0], nodeCount)
			mainResource.SubResources = subResources
		}
	}

	return mainResource
}

func aksOSDiskSubResource(region string, diskSize int, instanceType string) *schema.Resource {
	diskType := aksGetStorageType(instanceType)
	storageReplicationType := "LRS"

	diskName := mapDiskName(diskType, diskSize)
	if diskName == "" {
		logging.Logger.Warn().Msgf("Could not map disk type %s and size %d to disk name", diskType, diskSize)
		return nil
	}

	productName, ok := diskProductNameMap[diskType]
	if !ok {
		logging.Logger.Warn().Msgf("Could not map disk type %s to product name", diskType)
		return nil
	}

	costComponent := []*schema.CostComponent{storageCostComponent(region, diskName, storageReplicationType, productName)}

	return &schema.Resource{
		Name:           "os_disk",
		CostComponents: costComponent,
	}
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
