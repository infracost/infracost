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
		Name:  "compute.azure.crossplane.io/AKSCluster",
		RFunc: NewAzureRMKubernetesCluster,
	}
}

// NewAzureRMKubernetesCluster ...
// Reference: https://doc.crds.dev/github.com/crossplane/provider-azure/compute.azure.crossplane.io/AKSCluster/v1alpha3@v0.16.1
func NewAzureRMKubernetesCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{})
	var costComponents []*schema.CostComponent
	var subResources []*schema.Resource

	skuTier := "Free"
	if d.Get("sku_tier").Type != gjson.Null {
		skuTier = d.Get("sku_tier").String()
	}

	if strings.ToLower(skuTier) == "paid" {
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
	if d.Get("nodeCount").Type != gjson.Null {
		nodeCount = decimal.NewFromInt(d.Get("nodeCount").Int())
	}

	subResources = []*schema.Resource{
		aksClusterNodePool("default_node_pool", region, d.RawValues, nodeCount, u),
	}

	if d.Get("network_profile.0.load_balancer_sku").Type != gjson.Null {
		if strings.ToLower(d.Get("network_profile.0.load_balancer_sku").String()) == "standard" {
			// region = convertRegion(region)
			// var monthlyDataProcessedGb *decimal.Decimal
			// if u != nil && u.Get("load_balancer.monthly_data_processed_gb").Type != gjson.Null {
			// 	monthlyDataProcessedGb = decimalPtr(decimal.NewFromInt(u.Get("load_balancer.monthly_data_processed_gb").Int()))
			// }
			lbResource := schema.Resource{
				Name: "Load Balancer",
				// CostComponents: []*schema.CostComponent{dataProcessedCostComponent(region, monthlyDataProcessedGb)},
			}
			subResources = append(subResources, &lbResource)
		}
	}
	if d.Get("addon_profile.0.http_application_routing").Type != gjson.Null {
		if strings.ToLower(d.Get("addon_profile.0.http_application_routing.0.enabled").String()) == "true" {
			if strings.HasPrefix(strings.ToLower(region), "usgov") {
				region = "US Gov Zone 1"
			} else if strings.HasPrefix(strings.ToLower(region), "germany") {
				region = "DE Zone 1"
			} else if strings.HasPrefix(strings.ToLower(region), "china") {
				region = "Zone 1 (China)"
			} else {
				region = "Zone 1"
			}

			dnsResource := schema.Resource{
				Name: "DNS",
				// CostComponents: []*schema.CostComponent{hostedPublicZoneCostComponent(region)},
			}
			subResources = append(subResources, &dnsResource)
		}
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
		SubResources:   subResources,
	}
}
