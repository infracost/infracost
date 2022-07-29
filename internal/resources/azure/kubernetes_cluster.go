package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"

	"strings"

	"github.com/shopspring/decimal"
)

type KubernetesCluster struct {
	Address                       string
	Region                        string
	SKUTier                       string
	NetworkProfileLoadBalancerSKU string
	DefaultNodePoolNodeCount      int64
	DefaultNodePoolOSDiskType     string
	DefaultNodePoolVMSize         string
	DefaultNodePoolOSDiskSizeGB   int64
	HttpApplicationRoutingEnabled bool
	LoadBalancer                  *KubernetesClusterLoadBalancer    `infracost_usage:"load_balancer"`
	DefaultNodePool               *KubernetesClusterDefaultNodePool `infracost_usage:"default_node_pool"`
}

type KubernetesClusterLoadBalancer struct {
	MonthlyDataProcessedGB int64 `infracost_usage:"monthly_data_processed_gb"`
}

type KubernetesClusterDefaultNodePool struct {
	Nodes int64 `infracost_usage:"nodes"`
}

var KubernetesClusterLoadBalancerSchema = []*schema.UsageItem{{Key: "monthly_data_processed_gb", ValueType: schema.Int64, DefaultValue: 0}}

var KubernetesClusterDefaultNodePoolSchema = []*schema.UsageItem{{Key: "nodes", ValueType: schema.Int64, DefaultValue: 0}}

var KubernetesClusterUsageSchema = []*schema.UsageItem{
	{
		Key:          "load_balancer",
		ValueType:    schema.SubResourceUsage,
		DefaultValue: &usage.ResourceUsage{Name: "load_balancer", Items: KubernetesClusterLoadBalancerSchema},
	},
	{
		Key:          "default_node_pool",
		ValueType:    schema.SubResourceUsage,
		DefaultValue: &usage.ResourceUsage{Name: "default_node_pool", Items: KubernetesClusterDefaultNodePoolSchema},
	},
}

func (r *KubernetesCluster) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *KubernetesCluster) BuildResource() *schema.Resource {
	region := r.Region
	var costComponents []*schema.CostComponent
	var subResources []*schema.Resource

	skuTier := "Free"
	if r.SKUTier != "" {
		skuTier = r.SKUTier
	}

	// Azure switched from "Paid" to "Standard" in API version 2023-02-01
	// (Terraform Azure provider version v3.51.0)
	if contains([]string{"paid", "standard"}, strings.ToLower(skuTier)) {
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
					{Key: "meterName", Value: strPtr("Standard Uptime SLA")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("Consumption"),
			},
		})
	}

	nodeCount := decimal.NewFromInt(1)
	if r.DefaultNodePoolNodeCount > 0 {
		nodeCount = decimal.NewFromInt(r.DefaultNodePoolNodeCount)
	}
	if r.DefaultNodePool.Nodes > 0 {
		nodeCount = decimal.NewFromInt(r.DefaultNodePool.Nodes)
	}

	subResources = []*schema.Resource{
		aksClusterNodePool("default_node_pool", region, r.DefaultNodePoolVMSize, r.DefaultNodePoolOSDiskType, r.DefaultNodePoolOSDiskSizeGB, nodeCount),
	}

	if strings.ToLower(r.NetworkProfileLoadBalancerSKU) == "standard" {
		region = convertRegion(region)
		monthlyDataProcessedGB := decimalPtr(decimal.NewFromInt(r.LoadBalancer.MonthlyDataProcessedGB))
		lbResource := schema.Resource{
			Name:           "Load Balancer",
			CostComponents: []*schema.CostComponent{lbDataProcessedCostComponent(region, monthlyDataProcessedGB)}, UsageSchema: KubernetesClusterUsageSchema,
		}
		subResources = append(subResources, &lbResource)
	}

	if r.HttpApplicationRoutingEnabled {
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
			CostComponents: []*schema.CostComponent{hostedPublicZoneCostComponent(region)}, UsageSchema: KubernetesClusterUsageSchema,
		}
		subResources = append(subResources, &dnsResource)
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		SubResources:   subResources, UsageSchema: KubernetesClusterUsageSchema,
	}
}
