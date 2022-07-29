package azure

import (
<<<<<<< HEAD
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"

=======
	"github.com/infracost/infracost/internal/resources/azure"
>>>>>>> eb7adbfa... refactor(azure): migrate more resources
	"github.com/infracost/infracost/internal/schema"
)

func getAzureKubernetesClusterNodePoolRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_kubernetes_cluster_node_pool",
		RFunc: NewKubernetesClusterNodePool,
		ReferenceAttributes: []string{
			"kubernetes_cluster_id",
		},
	}
}
<<<<<<< HEAD

func NewAzureRMKubernetesClusterNodePool(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{"kubernetes_cluster_id"})

	nodeCount := decimal.NewFromInt(1)

	// if the node count is not set explicitly let's take the min_count.
	if d.Get("min_count").Type != gjson.Null {
		nodeCount = decimal.NewFromInt(d.Get("min_count").Int())
	}

	if d.Get("node_count").Type != gjson.Null {
		nodeCount = decimal.NewFromInt(d.Get("node_count").Int())
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

	var monthlyHours *float64
	if u != nil && u.Get("monthly_hrs").Exists() {
		monthlyHours = u.GetFloat("monthly_hrs")
	}

	os := "Linux"
	if n.Get("os_type").Type != gjson.Null {
		os = n.Get("os_type").String()
	}

	if n.Get("os_sku").Type != gjson.Null {
		if strings.HasPrefix(strings.ToLower(n.Get("os_sku").String()), "windows") {
			os = "Windows"
		}
	}

	if strings.EqualFold(os, "windows") {
		costComponents = append(costComponents, windowsVirtualMachineCostComponent(region, instanceType, "None", monthlyHours))
	} else {
		costComponents = append(costComponents, linuxVirtualMachineCostComponent(region, instanceType, monthlyHours))
	}

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
		log.Warn().Msgf("Could not map disk type %s and size %d to disk name", diskType, diskSize)
		return nil
	}

	productName, ok := diskProductNameMap[diskType]
	if !ok {
		log.Warn().Msgf("Could not map disk type %s to product name", diskType)
		return nil
	}

	costComponent := []*schema.CostComponent{storageCostComponent(region, diskName, storageReplicationType, productName)}

	return &schema.Resource{
		Name:           "os_disk",
		CostComponents: costComponent,
	}
=======
func NewKubernetesClusterNodePool(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &azure.KubernetesClusterNodePool{
		Address:      d.Address,
		Region:       lookupRegion(d, []string{"kubernetes_cluster_id"}),
		VMSize:       d.Get("vm_size").String(),
		OSDiskType:   d.Get("os_disk_type").String(),
		OSDiskSizeGB: d.Get("os_disk_size_gb").Int(),
		NodeCount:    d.Get("node_count").Int(),
	}
	r.PopulateUsage(u)
	return r.BuildResource()
>>>>>>> eb7adbfa... refactor(azure): migrate more resources
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
