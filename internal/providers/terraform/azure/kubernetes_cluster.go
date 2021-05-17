package azure

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetAzureRMKubernetesClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_kubernetes_cluster",
		RFunc: NewAzureRMKubernetesCluster,
	}
}

func NewAzureRMKubernetesCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var costComponents []*schema.CostComponent
	var subResources []*schema.Resource

	location := d.Get("location").String()

	skuTier := "Free"
	if d.Get("sku_tier").Type != gjson.Null {
		skuTier = d.Get("sku_tier").String()
	}

	if skuTier == "Paid" {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:           "Uptime SLA",
			Unit:           "hours",
			UnitMultiplier: 1,
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(location),
				Service:       strPtr("Azure Kubernetes Service"),
				ProductFamily: strPtr("Compute"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "skuName", Value: strPtr("Standard")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("Consumption"),
			},
		})
	}

	instanceType := d.Get("default_node_pool.0.vm_size").String()
	nodeCount := decimal.NewFromInt(1)
	if d.Get("default_node_pool.0.node_count").Type != gjson.Null {
		nodeCount = decimal.NewFromInt(d.Get("default_node_pool.0.node_count").Int())
	}
	if u != nil && u.Get("default_node_pool.nodes").Exists() {
		nodeCount = decimal.NewFromInt(u.Get("default_node_pool.nodes").Int())
	}

	subResources = append(subResources, &schema.Resource{
		Name:           "default_node_pool",
		CostComponents: []*schema.CostComponent{linuxVirtualMachineCostComponent(location, instanceType)},
	})
	schema.MultiplyQuantities(subResources[0], nodeCount)

	osDiskType := "Managed"
	if d.Get("default_node_pool.0.os_disk_type").Type != gjson.Null {
		osDiskType = d.Get("default_node_pool.0.os_disk_type").String()
	}
	if osDiskType == "Managed" {
		var diskSize int
		if d.Get("default_node_pool.0.os_disk_size_gb").Type != gjson.Null {
			diskSize = int(d.Get("default_node_pool.0.os_disk_size_gb").Int())
		}
		osDisk := aksOSDiskSubResource(location, diskSize, u)

		if osDisk != nil {
			subResources[0].SubResources = append(subResources[0].SubResources, osDisk)
			schema.MultiplyQuantities(subResources[0].SubResources[0], nodeCount)
		}
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
		SubResources:   subResources,
	}
}
