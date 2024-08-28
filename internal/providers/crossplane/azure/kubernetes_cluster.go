package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

// GetAzureRMKubernetesClusterRegistryItem ....
func GetAzureRMKubernetesClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "containerservice.azure.upbound.io/v1beta1",
		RFunc: NewAzureRMKubernetesCluster,
	}
}

// NewAzureRMKubernetesCluster ...
// Reference: https://marketplace.upbound.io/providers/upbound/provider-azure-containerservice/v1.5.0
func NewAzureRMKubernetesCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{})
	var costComponents []*schema.CostComponent
	var subResources []*schema.Resource

	forProvider := d.Get("forProvider")
    
	skuTier := "Free"
	if forProvider.Get("sku_tier").Type != gjson.Null {
		skuTier = strings.Title(strings.ToLower(forProvider.Get("sku_tier").String()))
		if skuTier != "Free" && skuTier != "Standard" && skuTier != "Premium" {
			skuTier = "Free" // Defaulting to Free if an unsupported value is provided
		}
	}

	if skuTier == "Standard" || skuTier == "Premium" {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:           "Uptime SLA",
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr(region),
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
	if forProvider.Get("defaultNodePool.0.nodeCount").Type != gjson.Null {
		nodeCount = decimal.NewFromInt(forProvider.Get("defaultNodePool.0.nodeCount").Int())
	}

	subResources = []*schema.Resource{
		aksClusterNodePool("default_node_pool", region, d.RawValues, nodeCount, u),
	}

	if forProvider.Get("networkProfile.0.loadBalancerSku").Type != gjson.Null {
		if strings.ToLower(forProvider.Get("networkProfile.0.loadBalancerSku").String()) == "standard" {
			lbResource := schema.Resource{
				Name: "Load Balancer",
			}
			subResources = append(subResources, &lbResource)
		}
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
		SubResources:   subResources,
	}
}
