package azure

import (
	"strings"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMKubernetesClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_kubernetes_cluster",
		RFunc: NewAzureRMKubernetesCluster,
	}
}

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
	if d.Get("default_node_pool.0.node_count").Type != gjson.Null {
		nodeCount = decimal.NewFromInt(d.Get("default_node_pool.0.node_count").Int())
	}

	// if the node count is not set explicitly let's take the min_count.
	if d.Get("default_node_pool.0.min_count").Type != gjson.Null && nodeCount.Equal(decimal.NewFromInt(1)) {
		nodeCount = decimal.NewFromInt(d.Get("default_node_pool.0.min_count").Int())
	}

	if u != nil {
		if v, ok := u.Get("default_node_pool").Map()["nodes"]; ok {
			nodeCount = decimal.NewFromInt(v.Int())
		}
	}

	subResources = []*schema.Resource{
		aksClusterNodePool("default_node_pool", region, d.Get("default_node_pool.0"), nodeCount, u),
	}

	if d.Get("network_profile.0.load_balancer_sku").Type != gjson.Null {
		if strings.ToLower(d.Get("network_profile.0.load_balancer_sku").String()) == "standard" {
			region = convertRegion(region)
			var monthlyDataProcessedGb *decimal.Decimal
			if u != nil {
				if v, ok := u.Get("load_balancer").Map()["monthly_data_processed_gb"]; ok {
					monthlyDataProcessedGb = decimalPtr(decimal.NewFromInt(v.Int()))
				}
			}
			lbResource := schema.Resource{
				Name:           "Load Balancer",
				CostComponents: []*schema.CostComponent{dataProcessedCostComponent(region, monthlyDataProcessedGb)},
			}
			subResources = append(subResources, &lbResource)
		}
	}

	routingEnabled := d.Get("http_application_routing_enabled").Bool()
	// Deprecated and removed in v3
	if d.Get("addon_profile.0.http_application_routing").Type != gjson.Null {
		routingEnabled = d.Get("addon_profile.0.http_application_routing.0.enabled").Bool()
	}
	if routingEnabled {
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
			Name:           "DNS",
			CostComponents: []*schema.CostComponent{hostedPublicZoneCostComponent(region)},
		}
		subResources = append(subResources, &dnsResource)
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
		SubResources:   subResources,
	}
}
