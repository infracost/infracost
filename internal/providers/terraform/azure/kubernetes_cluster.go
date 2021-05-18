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

	nodeCount := decimal.NewFromInt(1)
	if d.Get("default_node_pool.0.node_count").Type != gjson.Null {
		nodeCount = decimal.NewFromInt(d.Get("default_node_pool.0.node_count").Int())
	}
	if u != nil && u.Get("default_node_pool.nodes").Exists() {
		nodeCount = decimal.NewFromInt(u.Get("default_node_pool.nodes").Int())
	}

	subResources = []*schema.Resource{
		aksClusterNodePool("default_node_pool", location, d.Get("default_node_pool.0"), nodeCount, u),
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
		SubResources:   subResources,
	}
}
